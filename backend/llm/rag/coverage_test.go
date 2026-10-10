package rag

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	qt "github.com/valyala/quicktemplate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

func requireEmbeddingStub(t *testing.T) {
	t.Helper()
	if os.Getenv("INFINITY_TIERS") == "" {
		t.Skip("INFINITY_TIERS is not set: no embedding provider is registered in this process")
	}
}

func useKnowledge(t *testing.T, db *gorm.DB) {
	t.Helper()
	saved := database.Knowledge
	database.Knowledge = db
	t.Cleanup(func() { database.Knowledge = saved })
}

func sqliteKnowledge(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Goal{}, &model.Methodology{}, &model.Modifier{},
		&model.Equipment{}, &model.Muscle{}, &model.Exercise{},
	); err != nil {
		t.Fatalf("migrate knowledge: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
	return db
}

func seedKnowledge(t *testing.T, db *gorm.DB) {
	t.Helper()
	strength := model.Methodology{ID: "strength", Name: "Strength", Description: "progressive overload"}
	if err := strength.SetWork(model.MethodologyWork{MinDifficulty: 20, MaxDifficulty: 90}); err != nil {
		t.Fatalf("set work: %v", err)
	}
	circuit := model.Methodology{ID: "circuit", Name: "Circuit", Description: "rounds"}
	if err := circuit.SetWork(model.MethodologyWork{}); err != nil {
		t.Fatalf("set work: %v", err)
	}
	rows := []any{
		&model.Goal{ID: "build-muscle", Description: "hypertrophy"},
		&model.Goal{ID: "lose-fat", Description: "fat loss"},
		&strength,
		&circuit,
		&model.Modifier{ID: "band", Patterns: []string{"row"}},
		&model.Equipment{ID: "barbell"},
		&model.Equipment{ID: "bench"},
		&model.Muscle{ID: "quads"},
		&model.Muscle{ID: "core"},
		&model.Exercise{ID: "squat", Name: "Squat", Muscles: []string{"quads"}, Equipment: []string{"barbell"}, Difficulty: 50, Mode: "reps"},
		&model.Exercise{ID: "hip-stretch", Name: "Hip Stretch", Muscles: []string{"quads"}, Difficulty: 10, IsMobility: true, Mode: "duration"},
		&model.Exercise{ID: "cat-cow", Name: "Cat Cow", Muscles: []string{"core"}, Difficulty: 5, IsMobility: true, Mode: "duration"},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed %T: %v", row, err)
		}
	}
}

// fakeKnowledgeDB is a gorm DB on a fake postgres driver whose query
// answers are canned per SQL shape: similarity queries over the
// embedding tables return fixed exercise and fact rows, everything else
// returns an empty result or the fixed exercise catalog.
type fakeKnowledgeConnector struct {
	failOn string
}

func (c *fakeKnowledgeConnector) Connect(context.Context) (driver.Conn, error) {
	return &fakeKnowledgeConn{connector: c}, nil
}

func (c *fakeKnowledgeConnector) Driver() driver.Driver { return fakeKnowledgeDriver{} }

type fakeKnowledgeDriver struct{}

func (fakeKnowledgeDriver) Open(string) (driver.Conn, error) {
	return &fakeKnowledgeConn{connector: &fakeKnowledgeConnector{}}, nil
}

type fakeKnowledgeConn struct {
	connector *fakeKnowledgeConnector
}

func (c *fakeKnowledgeConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeKnowledgeStmt{conn: c, query: query}, nil
}

func (c *fakeKnowledgeConn) Close() error { return nil }

func (c *fakeKnowledgeConn) Begin() (driver.Tx, error) { return fakeKnowledgeTx{}, nil }

func (c *fakeKnowledgeConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	if c.connector.failOn != "" && strings.Contains(query, c.connector.failOn) {
		return nil, errors.New("fake exec failure")
	}
	return driver.RowsAffected(1), nil
}

var exerciseColumns = []string{
	"exercise_id", "text", "distance", "has_equipment",
	"id", "name", "aliases", "equipment", "muscles", "reference",
	"instructions", "cues", "difficulty", "is_mobility", "mode",
	"created_at", "updated_at",
}

func exerciseRow(id, name, muscles string, difficulty int64, mobility bool) []driver.Value {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return []driver.Value{
		id, "text for " + id, 0.05, false,
		id, name, "{}", "{}", muscles, "",
		"{}", "{}", difficulty, mobility, "reps",
		now, now,
	}
}

func (c *fakeKnowledgeConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	if c.connector.failOn != "" && strings.Contains(query, c.connector.failOn) {
		return nil, errors.New("fake query failure")
	}
	lower := strings.ToLower(query)
	switch {
	case strings.Contains(lower, "fact_embeddings"):
		columns := []string{"fact_id", "text", "distance", "id", "reference", "area", "tags", "content", "created_at", "updated_at"}
		if strings.Contains(lower, "keyword_rank") {
			columns = []string{"fact_id", "text", "distance", "keyword_rank", "id", "reference", "area", "tags", "content", "created_at", "updated_at"}
			row := []driver.Value{"f1", "fact text", 0.1, 0.9, "00000000-0000-0000-0000-0000000000f1", "ref-1", "recovery", "{sleep}", "sleep well", time.Now(), time.Now()}
			return &fakeKnowledgeRows{columns: columns, values: [][]driver.Value{row}}, nil
		}
		row := []driver.Value{"f1", "fact text", 0.1, "00000000-0000-0000-0000-0000000000f1", "ref-1", "recovery", "{sleep}", "sleep well", time.Now(), time.Now()}
		return &fakeKnowledgeRows{columns: columns, values: [][]driver.Value{row}}, nil
	case strings.Contains(lower, "exercise_embeddings"):
		columns := exerciseColumns
		if strings.Contains(lower, "keyword_rank") {
			columns = []string{
				"exercise_id", "text", "distance", "keyword_rank", "has_equipment",
				"id", "name", "aliases", "equipment", "muscles", "reference",
				"instructions", "cues", "difficulty", "is_mobility", "mode",
				"created_at", "updated_at",
			}
			row := exerciseRow("squat", "Squat", "{quads}", 50, false)
			row = append(row[:3], append([]driver.Value{0.7}, row[3:]...)...)
			return &fakeKnowledgeRows{columns: columns, values: [][]driver.Value{row}}, nil
		}
		return &fakeKnowledgeRows{columns: columns, values: [][]driver.Value{
			exerciseRow("squat", "Squat", "{quads}", 50, false),
			exerciseRow("deadlift", "Deadlift", "{back}", 70, false),
			exerciseRow("hip-stretch", "Hip Stretch", "{quads}", 10, true),
		}}, nil
	case strings.Contains(lower, "exercises"):
		columns := []string{
			"id", "name", "aliases", "equipment", "muscles", "reference",
			"instructions", "cues", "difficulty", "is_mobility", "mode", "created_at", "updated_at",
		}
		trim := func(row []driver.Value) []driver.Value { return row[4:] }
		return &fakeKnowledgeRows{columns: columns, values: [][]driver.Value{
			trim(exerciseRow("squat", "Squat", "{quads}", 50, false)),
			trim(exerciseRow("hip-stretch", "Hip Stretch", "{quads}", 10, true)),
			trim(exerciseRow("cat-cow", "Cat Cow", "{core}", 5, true)),
		}}, nil
	default:
		return &fakeKnowledgeRows{}, nil
	}
}

type fakeKnowledgeTx struct{}

func (fakeKnowledgeTx) Commit() error   { return nil }
func (fakeKnowledgeTx) Rollback() error { return nil }

type fakeKnowledgeStmt struct {
	conn  *fakeKnowledgeConn
	query string
}

func (s *fakeKnowledgeStmt) Close() error  { return nil }
func (s *fakeKnowledgeStmt) NumInput() int { return -1 }

func (s *fakeKnowledgeStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return s.conn.ExecContext(context.Background(), s.query, nil)
}

func (s *fakeKnowledgeStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return s.conn.QueryContext(context.Background(), s.query, nil)
}

type fakeKnowledgeRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *fakeKnowledgeRows) Columns() []string { return r.columns }
func (r *fakeKnowledgeRows) Close() error      { return nil }
func (r *fakeKnowledgeRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func fakeKnowledge(t *testing.T, connector *fakeKnowledgeConnector) *gorm.DB {
	t.Helper()
	sqlDB := sql.OpenDB(connector)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open fake knowledge: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func TestSimpleRetrieversOnSQLite(t *testing.T) {
	db := sqliteKnowledge(t)
	seedKnowledge(t, db)
	useKnowledge(t, db)

	goals, err := RetrieveGoals([]string{"build-muscle", "lose-fat"})
	if err != nil || len(goals) != 2 {
		t.Errorf("RetrieveGoals = %v, %v", goals, err)
	}
	if goals, err := RetrieveGoals(nil); goals != nil || err != nil {
		t.Errorf("RetrieveGoals(nil) = %v, %v", goals, err)
	}

	methodology, err := RetrieveMethodology("strength")
	if err != nil || methodology == nil || methodology.GetWork().MinDifficulty != 20 {
		t.Errorf("RetrieveMethodology = %v, %v", methodology, err)
	}
	if _, err := RetrieveMethodology("absent"); err == nil {
		t.Error("missing methodology must error")
	}
	if methodology, err := RetrieveMethodology(""); methodology != nil || err != nil {
		t.Errorf("RetrieveMethodology(empty) = %v, %v", methodology, err)
	}

	all, err := RetrieveAllMethodologies()
	if err != nil || len(all) != 2 {
		t.Errorf("RetrieveAllMethodologies = %v, %v", all, err)
	}

	modifiers, err := RetrieveUserModifiers([]string{"band"})
	if err != nil || len(modifiers) != 1 {
		t.Errorf("RetrieveUserModifiers = %v, %v", modifiers, err)
	}
	if modifiers, err := RetrieveUserModifiers(nil); modifiers != nil || err != nil {
		t.Errorf("RetrieveUserModifiers(nil) = %v, %v", modifiers, err)
	}

	equipment, err := RetrieveEquipment([]string{"barbell", "bench"})
	if err != nil || len(equipment) != 2 {
		t.Errorf("RetrieveEquipment = %v, %v", equipment, err)
	}
	if equipment, err := RetrieveEquipment(nil); equipment != nil || err != nil {
		t.Errorf("RetrieveEquipment(nil) = %v, %v", equipment, err)
	}

	warmup, err := RetrieveWarmupExercises()
	if err != nil || len(warmup) != 2 {
		t.Errorf("RetrieveWarmupExercises = %d exercises, %v", len(warmup), err)
	}

	if _, err := RetrieveCooldownExercises([]string{"quads", "core", "quads"}); err != nil {
		t.Errorf("RetrieveCooldownExercises: %v", err)
	}
	cooldownAll, err := RetrieveCooldownExercises(nil)
	if err != nil {
		t.Errorf("RetrieveCooldownExercises(nil) = %v, %v", cooldownAll, err)
	}
}

func TestSimpleRetrieverErrors(t *testing.T) {
	db := fakeKnowledge(t, &fakeKnowledgeConnector{failOn: "goals"})
	useKnowledge(t, db)
	if _, err := RetrieveGoals([]string{"x"}); err == nil {
		t.Error("RetrieveGoals must surface query errors")
	}

	db = fakeKnowledge(t, &fakeKnowledgeConnector{failOn: "methodologies"})
	useKnowledge(t, db)
	if _, err := RetrieveAllMethodologies(); err == nil {
		t.Error("RetrieveAllMethodologies must surface query errors")
	}

	db = fakeKnowledge(t, &fakeKnowledgeConnector{failOn: "modifiers"})
	useKnowledge(t, db)
	if _, err := RetrieveUserModifiers([]string{"band"}); err == nil {
		t.Error("RetrieveUserModifiers must surface query errors")
	}

	db = fakeKnowledge(t, &fakeKnowledgeConnector{failOn: "equipment"})
	useKnowledge(t, db)
	if _, err := RetrieveEquipment([]string{"barbell"}); err == nil {
		t.Error("RetrieveEquipment must surface query errors")
	}

	db = fakeKnowledge(t, &fakeKnowledgeConnector{failOn: "exercises"})
	useKnowledge(t, db)
	if _, err := RetrieveWarmupExercises(); err == nil {
		t.Error("RetrieveWarmupExercises must surface query errors")
	}
	if _, err := RetrieveCooldownExercises([]string{"quads"}); err != nil {
		t.Errorf("cooldown per-muscle errors are tolerated, got %v", err)
	}
	if _, err := RetrieveCooldownExercises(nil); err != nil {
		t.Errorf("RetrieveCooldownExercises with an empty muscle table = %v", err)
	}

	db = fakeKnowledge(t, &fakeKnowledgeConnector{failOn: "muscles"})
	useKnowledge(t, db)
	if _, err := RetrieveCooldownExercises(nil); err == nil {
		t.Error("RetrieveCooldownExercises must surface the muscle listing error")
	}
	if _, err := RetrieveWorkExercises(
		[]model.Profile{{}}, nil, nil, nil, 20, nil, nil, "", nil, nil, nil, 30,
	); err == nil {
		t.Error("RetrieveWorkExercises must surface the muscle listing error")
	}
}

func TestSimilarityRetrievalOnFakePostgres(t *testing.T) {
	requireEmbeddingStub(t)
	useKnowledge(t, fakeKnowledge(t, &fakeKnowledgeConnector{}))

	profiles := []model.Profile{{Gender: "female"}}
	work, err := RetrieveWorkExercises(
		profiles, []string{"build-muscle"}, []string{"barbell"},
		map[string]float64{"quads": 60}, 25,
		&model.Methodology{ID: "strength"},
		[]string{"quads"}, "hard session",
		[]string{"squat"}, nil, nil, 45,
	)
	if err != nil || len(work) == 0 {
		t.Errorf("RetrieveWorkExercises = %d exercises, %v", len(work), err)
	}

	facts, err := RetrieveUserFacts(profiles, []string{"build-muscle"}, "sleep")
	if err != nil || len(facts) != 1 {
		t.Errorf("RetrieveUserFacts hybrid = %v, %v", facts, err)
	}
	facts, err = RetrieveUserFacts(profiles, nil, "")
	if err != nil || len(facts) != 1 {
		t.Errorf("RetrieveUserFacts pure vector = %v, %v", facts, err)
	}

	favorites, err := RetrieveFavoriteExercises([]string{"squat"})
	if err != nil || len(favorites) == 0 {
		t.Errorf("RetrieveFavoriteExercises = %v, %v", favorites, err)
	}
	if favorites, err := RetrieveFavoriteExercises(nil); favorites != nil || err != nil {
		t.Errorf("RetrieveFavoriteExercises(nil) = %v, %v", favorites, err)
	}

	pool, pins, err := RetrieveExplicitProgramExercises([]string{"squat", "hip stretch", "nonexistent movement xyz"})
	if err != nil {
		t.Fatalf("RetrieveExplicitProgramExercises: %v", err)
	}
	if len(pins) == 0 || len(pool) < len(pins) {
		t.Errorf("pool = %d, pins = %d", len(pool), len(pins))
	}
	if pool, pins, err := RetrieveExplicitProgramExercises(nil); pool != nil || pins != nil || err != nil {
		t.Errorf("RetrieveExplicitProgramExercises(nil) = %v, %v, %v", pool, pins, err)
	}
}

func TestSimilarityRetrievalErrors(t *testing.T) {
	requireEmbeddingStub(t)
	useKnowledge(t, fakeKnowledge(t, &fakeKnowledgeConnector{failOn: "embeddings"}))

	if _, err := retrieveBySimilarity([]float32{1, 2}, "", nil, []string{"quads"}, model.MethodologyWork{}, nil, 10, false); err == nil {
		t.Error("retrieveBySimilarity must surface query errors")
	}
	if _, err := retrieveBySimilarity([]float32{1, 2}, "squat quads", []string{"barbell"}, nil, model.MethodologyWork{MobilityOnly: true}, []string{"x"}, 10, false); err == nil {
		t.Error("retrieveBySimilarity with keywords must surface query errors")
	}
	if _, err := RetrieveUserFacts([]model.Profile{{}}, []string{"g"}, ""); err == nil {
		t.Error("RetrieveUserFacts must surface query errors")
	}
	if _, err := RetrieveFavoriteExercises([]string{"squat"}); err == nil {
		t.Error("RetrieveFavoriteExercises must surface query errors")
	}
	pool, _, err := RetrieveExplicitProgramExercises([]string{"squat"})
	if err != nil || len(pool) == 0 {
		t.Errorf("explicit program neighbor failures are tolerated: pool = %d, err = %v", len(pool), err)
	}

	// balanced retrieval tolerates per-muscle failures.
	got := retrieveBalancedByMuscle([]float32{1}, "", nil, nil, []string{"quads"}, model.MethodologyWork{}, nil, 20, nil, nil, nil, 30)
	if len(got) != 0 {
		t.Errorf("retrieveBalancedByMuscle with failing queries = %v", got)
	}
	if got := retrieveBalancedByMuscle(nil, "", nil, nil, nil, model.MethodologyWork{}, nil, 20, nil, nil, nil, 30); got != nil {
		t.Errorf("retrieveBalancedByMuscle without muscles = %v", got)
	}
}

func TestWorkExercisesMusclesFromDB(t *testing.T) {
	requireEmbeddingStub(t)
	db := sqliteKnowledge(t)
	seedKnowledge(t, db)
	useKnowledge(t, db)

	// the similarity SQL is postgres-only, so on sqlite every muscle
	// lookup fails and the balanced retrieval degrades to an empty pool
	// with no error.
	exercises, err := RetrieveWorkExercises(
		[]model.Profile{{}}, nil, nil, nil, 20, nil, nil, "", nil, nil, nil, 30,
	)
	if err != nil || len(exercises) != 0 {
		t.Errorf("RetrieveWorkExercises on sqlite = %d, %v", len(exercises), err)
	}
}

func TestScalingHelpers(t *testing.T) {
	for duration, want := range map[int]int{20: 16, 30: 16, 45: 20, 60: 22, 90: 30, 120: 40} {
		if got := maxWorkExercises(duration); got != want {
			t.Errorf("maxWorkExercises(%d) = %d, want %d", duration, got, want)
		}
	}
	for duration, want := range map[int]int{30: 4, 45: 4, 90: 5, 120: 7} {
		if got := maxPerMuscleExercises(duration); got != want {
			t.Errorf("maxPerMuscleExercises(%d) = %d, want %d", duration, got, want)
		}
	}
	if got := buildKeywordQuery(nil, nil, nil, ""); got != "" {
		t.Errorf("buildKeywordQuery empty = %q", got)
	}
	if got := buildKeywordQuery([]string{"g"}, []string{"e"}, []string{"m"}, "two words"); got != "g e m two words" {
		t.Errorf("buildKeywordQuery = %q", got)
	}
}

func TestGenTemplates(t *testing.T) {
	profileData := `{"goals":["strength"],"injuries":[{"description":"knee","year":2020}],"limitations":["no jumping"],"conditions":["asthma"]}`
	profile := model.Profile{Gender: "female"}
	profile.Data = []byte(profileData)

	full := GenProfile([]model.Profile{profile}, nil, []string{"barbell"}, []string{"quads"}, "morning session")
	for _, want := range []string{"strength", "knee", "barbell", "quads", "morning session"} {
		if !strings.Contains(full, want) {
			t.Errorf("GenProfile missing %q in %q", want, full)
		}
	}
	if got := GenProfile(nil, []string{"g"}, nil, nil, ""); !strings.Contains(got, "Goals: g") {
		t.Errorf("GenProfile explicit goals = %q", got)
	}
	if got := GenProfile(nil, nil, nil, nil, ""); strings.TrimSpace(got) != "" {
		t.Errorf("GenProfile empty = %q", got)
	}

	exercise := model.Exercise{
		Name:         "Squat",
		Equipment:    []string{"barbell"},
		Muscles:      []string{"quads"},
		Instructions: []string{"brace", "descend"},
	}
	if got := GenExercise(exercise); !strings.Contains(got, "Squat") || !strings.Contains(got, "1. brace") {
		t.Errorf("GenExercise = %q", got)
	}
	if got := GenExercise(model.Exercise{Name: "Plank"}); !strings.Contains(got, "Plank") {
		t.Errorf("GenExercise bare = %q", got)
	}

	fact := model.Fact{Area: "recovery", Content: "sleep matters", Tags: []string{"sleep"}}
	if got := GenFact(fact); !strings.Contains(got, "recovery: sleep matters") || !strings.Contains(got, "Tags: sleep") {
		t.Errorf("GenFact = %q", got)
	}
	if got := GenFact(model.Fact{Area: "form", Content: "brace hard"}); strings.Contains(got, "Tags") {
		t.Errorf("GenFact without tags = %q", got)
	}

	var buf bytes.Buffer
	w := qt.AcquireWriter(&buf)
	defer qt.ReleaseWriter(w)
	StreamGenProfile(w, []model.Profile{profile}, []string{"g"}, nil, nil, "")
	StreamGenExercise(w, exercise)
	StreamGenFact(w, fact)
	WriteGenProfile(&buf, nil, nil, nil, nil, "x")
	WriteGenExercise(&buf, exercise)
	WriteGenFact(&buf, fact)
	if buf.Len() == 0 {
		t.Error("stream and write variants produced no output")
	}
}
