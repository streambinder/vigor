package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestMain starts the embedding stub the bootstrap calls need. The
// address matches the INFINITY_TIERS value CI exports for every Go
// test process; another package's identical stub may already own it.
// soReusePort is SO_REUSEPORT on Linux (the syscall package does
// not export the constant; golang.org/x/sys/unix would, but the
// tests stay on the standard library).
const soReusePort = 15

// listenStubPort opens a listener with SO_REUSEPORT. Every test
// binary that needs a stub owns its own listener for its whole
// lifetime; the kernel load-balances between the identical stubs
// of parallel test processes, and one process exiting never takes
// the port away from the others.
func listenStubPort(addr string) (net.Listener, error) {
	lc := net.ListenConfig{Control: func(_, _ string, c syscall.RawConn) error {
		var setErr error
		if err := c.Control(func(fd uintptr) {
			setErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, soReusePort, 1)
		}); err != nil {
			return err
		}
		return setErr
	}}
	return lc.Listen(context.Background(), "tcp", addr)
}

func TestMain(m *testing.M) {
	listener, err := listenStubPort("127.0.0.1:18741")
	if err == nil {
		mux := http.NewServeMux()
		mux.HandleFunc("/", knowledgeEmbeddingStub)
		server := &http.Server{Handler: mux}
		go func() { _ = server.Serve(listener) }()
		defer func() { _ = server.Close() }()
	}
	os.Exit(m.Run())
}

// knowledgeEmbeddingStub returns one deterministic vector per input,
// in the OpenAI-compatible shape the infinity provider parses.
func knowledgeEmbeddingStub(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Input []string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data := make([]map[string]any, 0, len(req.Input))
	for i := range req.Input {
		data = append(data, map[string]any{
			"embedding": []float32{float32(i + 1), 0.5, 0.25},
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Connection", "close")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

// openKnowledgeDB returns a gorm DB speaking to a fresh fake wire
// server, plus the server itself for statement assertions.
func openKnowledgeDB(t *testing.T) (*gorm.DB, *fakePGServer) {
	t.Helper()
	server := startFakePGServer(t)
	db, err := gorm.Open(postgres.Open(server.dsn("knowledge")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db, server
}

// packageDir is the knowledge module directory, captured before
// any test moves the working directory.
var packageDir, _ = os.Getwd()

// chdirFeatures moves the test into a scratch directory whose
// features folder holds the given files (name → content); empty
// content copies the real catalog file.
func chdirFeatures(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "features"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if content == "" {
			raw, err := os.ReadFile(filepath.Join(packageDir, "features", name))
			if err != nil {
				t.Fatal(err)
			}
			content = string(raw)
		}
		if err := os.WriteFile(filepath.Join(dir, "features", name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(packageDir) })
	return dir
}

func TestBootstrapCatalogsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)

	if err := bootstrapMethodologies(db); err != nil {
		t.Fatalf("bootstrapMethodologies: %v", err)
	}
	if err := bootstrapMuscles(db); err != nil {
		t.Fatalf("bootstrapMuscles: %v", err)
	}
	if err := bootstrapEquipment(db); err != nil {
		t.Fatalf("bootstrapEquipment: %v", err)
	}
	if err := bootstrapGoals(db); err != nil {
		t.Fatalf("bootstrapGoals: %v", err)
	}
	if err := bootstrapModifiers(db); err != nil {
		t.Fatalf("bootstrapModifiers: %v", err)
	}
	for _, table := range []string{
		"methodologies", "muscles",
		"equipment", "equipment_embeddings",
		"goals", "goal_embeddings",
		"modifiers", "modifier_embeddings",
	} {
		if !server.saw(`insert into "` + table + `"`) {
			t.Fatalf("no insert into %s recorded", table)
		}
	}
}

func TestBootstrapExercisesCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	if err := bootstrapEquipment(db); err != nil {
		t.Fatalf("bootstrapEquipment: %v", err)
	}
	if err := boostrapExercises(db); err != nil {
		t.Fatalf("boostrapExercises: %v", err)
	}
	if !server.saw(`insert into "exercises"`) {
		t.Fatal("no exercise inserts recorded")
	}
	if !server.saw(`insert into "exercise_embeddings"`) {
		t.Fatal("no exercise embedding inserts recorded")
	}
}

func TestBootstrapFactsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	if err := boostrapFacts(db); err != nil {
		t.Fatalf("boostrapFacts: %v", err)
	}
	if !server.saw(`insert into "facts"`) {
		t.Fatal("no fact inserts recorded")
	}
	if !server.saw(`insert into "fact_embeddings"`) {
		t.Fatal("no fact embedding inserts recorded")
	}
}

func TestBootstrapFileErrorsCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)

	// a directory without the catalog files fails every bootstrap
	// at the read step
	chdirFeatures(t, map[string]string{})
	for name, fn := range map[string]func(*gorm.DB) error{
		"methodologies": bootstrapMethodologies,
		"muscles":       bootstrapMuscles,
		"equipment":     bootstrapEquipment,
		"goals":         bootstrapGoals,
		"modifiers":     bootstrapModifiers,
		"exercises":     boostrapExercises,
		"facts":         boostrapFacts,
	} {
		if err := fn(db); err == nil {
			t.Fatalf("%s must fail without its catalog file", name)
		}
	}
}

func TestBootstrapBadJSONCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	chdirFeatures(t, map[string]string{
		"methodologies.json": "{bad",
		"muscles.json":       "{bad",
		"equipment.json":     "{bad",
		"goals.json":         "{bad",
		"modifiers.json":     "{bad",
		"exercises.json":     "{bad",
		"facts.json":         "{bad",
	})
	for name, fn := range map[string]func(*gorm.DB) error{
		"methodologies": bootstrapMethodologies,
		"muscles":       bootstrapMuscles,
		"equipment":     bootstrapEquipment,
		"goals":         bootstrapGoals,
		"modifiers":     bootstrapModifiers,
		"exercises":     boostrapExercises,
		"facts":         boostrapFacts,
	} {
		if err := fn(db); err == nil {
			t.Fatalf("%s must fail on malformed JSON", name)
		}
	}
}

func TestBootstrapStoreErrorsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	server.failOn = "INSERT INTO"
	for name, fn := range map[string]func(*gorm.DB) error{
		"methodologies": bootstrapMethodologies,
		"muscles":       bootstrapMuscles,
		"equipment":     bootstrapEquipment,
		"goals":         bootstrapGoals,
		"modifiers":     bootstrapModifiers,
		"exercises":     boostrapExercises,
		"facts":         boostrapFacts,
	} {
		if err := fn(db); err == nil {
			t.Fatalf("%s must fail when inserts fail", name)
		}
	}
}

func TestCreateVectorIndexesCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	if err := createVectorIndexes(db); err != nil {
		t.Fatalf("createVectorIndexes: %v", err)
	}
	if !server.saw("CREATE INDEX") {
		t.Fatal("no index DDL recorded")
	}

	db, server = openKnowledgeDB(t)
	server.failOn = "CREATE INDEX"
	if err := createVectorIndexes(db); err == nil {
		t.Fatal("createVectorIndexes must fail when the DDL fails")
	}
}

// The embedding cleanup in boostrapExercises returns the delete
// error when the wire refuses the statement. The other bootstraps
// do not check the cleanup result; their unchecked calls are
// production behavior, not a test gap.
func TestBootstrapDeleteErrorsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	server.failOn = "DELETE FROM"
	if err := boostrapExercises(db); err == nil {
		t.Fatal("boostrapExercises must fail when the embedding cleanup fails")
	}
}

// Replacing the exercise equipment association returns the wire
// error from the join table statements.
func TestBootstrapAssociationErrorCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	server.failOn = "exercise_equipment"
	if err := boostrapExercises(db); err == nil {
		t.Fatal("boostrapExercises must fail when the association write fails")
	}
}

// writeAnimations fills the animations folder of the scratch
// directory chdirFeatures prepared with the given files.
func writeAnimations(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "animations"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, "animations", name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBootstrapAnimationsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeAnimations(t, dir, map[string]string{
		"barbell-full-squat.json": `{"schema":"vigor-exercise-animation/1"}`,
		"burpee.json":             `{"schema":"vigor-exercise-animation/1"}`,
		"notes.txt":               "not an animation",
	})
	if err := bootstrapAnimations(db); err != nil {
		t.Fatalf("bootstrapAnimations: %v", err)
	}
	if !server.saw(`insert into "exercise_animations"`) {
		t.Fatal("no animation inserts recorded")
	}
}

// Without the animations directory the bootstrap skips the step:
// the catalog ships without animations, not with a failed build.
func TestBootstrapAnimationsMissingDirCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	chdirFeatures(t, map[string]string{})
	if err := bootstrapAnimations(db); err != nil {
		t.Fatalf("bootstrapAnimations without directory: %v", err)
	}
	if server.saw(`insert into "exercise_animations"`) {
		t.Fatal("animation inserts recorded without a directory")
	}
}

func TestBootstrapAnimationsBadJSONCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeAnimations(t, dir, map[string]string{"burpee.json": "{bad"})
	if err := bootstrapAnimations(db); err == nil {
		t.Fatal("bootstrapAnimations must fail on a malformed document")
	}
}

func TestBootstrapAnimationsStoreErrorCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeAnimations(t, dir, map[string]string{
		"burpee.json": `{"schema":"vigor-exercise-animation/1"}`,
	})
	server.failOn = "INSERT INTO"
	if err := bootstrapAnimations(db); err == nil {
		t.Fatal("bootstrapAnimations must fail when inserts fail")
	}
}

// testMannequinManifest is the manifest the models tests use: the real
// mannequin GLB and licence file, copied from the repository.
const testMannequinManifest = `{"models":[{"id":"quaternius-mannequin","displayName":"Quaternius Mannequin","licence":"CC0 1.0 Universal","source":"Universal Animation Library [Standard] (Quaternius)","file":"quaternius-mannequin.glb","licenceFile":"LICENSE-quaternius.txt"}]}`

// writeModels fills the models folder of the scratch directory
// chdirFeatures prepared with the given manifest and files.
func writeModels(t *testing.T, dir, manifest string, files map[string][]byte) {
	t.Helper()
	modelsDir := filepath.Join(dir, "features", "models")
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modelsDir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(modelsDir, name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// readPackageModel reads a committed file from features/models.
func readPackageModel(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(packageDir, "features", "models", name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// buildTestGLB frames the given node names as a minimal GLB file.
func buildTestGLB(t *testing.T, nodeNames []string) []byte {
	t.Helper()
	nodes := make([]map[string]string, 0, len(nodeNames))
	for _, name := range nodeNames {
		nodes = append(nodes, map[string]string{"name": name})
	}
	doc, err := json.Marshal(map[string]any{"asset": map[string]string{"version": "2.0"}, "nodes": nodes})
	if err != nil {
		t.Fatal(err)
	}
	for len(doc)%4 != 0 {
		doc = append(doc, ' ')
	}
	out := make([]byte, 0, 20+len(doc))
	out = append(out, []byte("glTF")...)
	version := make([]byte, 4)
	binary.LittleEndian.PutUint32(version, 2)
	out = append(out, version...)
	total := make([]byte, 4)
	binary.LittleEndian.PutUint32(total, uint32(20+len(doc)))
	out = append(out, total...)
	chunkLen := make([]byte, 4)
	binary.LittleEndian.PutUint32(chunkLen, uint32(len(doc)))
	out = append(out, chunkLen...)
	out = append(out, 0x4A, 0x53, 0x4F, 0x4E)
	out = append(out, doc...)
	return out
}

// writeStills fills the stills folder of the scratch directory
// chdirFeatures prepared with the given files.
func writeStills(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "stills"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, "stills", name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBootstrapModelsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeModels(t, dir, testMannequinManifest, map[string][]byte{
		"quaternius-mannequin.glb": readPackageModel(t, "quaternius-mannequin.glb"),
		"LICENSE-quaternius.txt":   readPackageModel(t, "LICENSE-quaternius.txt"),
	})
	if err := bootstrapModels(db); err != nil {
		t.Fatalf("bootstrapModels: %v", err)
	}
	if !server.saw(`insert into "exercise_models"`) {
		t.Fatal("no model inserts recorded")
	}
}

func TestBootstrapModelsMissingDirCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	chdirFeatures(t, map[string]string{})
	if err := bootstrapModels(db); err != nil {
		t.Fatalf("bootstrapModels without manifest: %v", err)
	}
	if server.saw(`insert into "exercise_models"`) {
		t.Fatal("model inserts recorded without a manifest")
	}
}

func TestBootstrapModelsBadManifestCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeModels(t, dir, "{bad", map[string][]byte{})
	if err := bootstrapModels(db); err == nil {
		t.Fatal("bootstrapModels must fail on a malformed manifest")
	}
}

func TestBootstrapModelsEntryErrorCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeModels(t, dir, `{"models":[{"id":"model-a"}]}`, map[string][]byte{})
	if err := bootstrapModels(db); err == nil {
		t.Fatal("bootstrapModels must fail on a manifest entry without a file")
	}
	dir = chdirFeatures(t, map[string]string{})
	writeModels(t, dir, `{"models":[{"file":"model-a.glb"}]}`, map[string][]byte{})
	if err := bootstrapModels(db); err == nil {
		t.Fatal("bootstrapModels must fail on a manifest entry without an id")
	}
}

func TestBootstrapModelsLicenceMissingCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeModels(t, dir, testMannequinManifest, map[string][]byte{
		"quaternius-mannequin.glb": readPackageModel(t, "quaternius-mannequin.glb"),
	})
	if err := bootstrapModels(db); err == nil {
		t.Fatal("bootstrapModels must fail when the licence file is missing")
	}
}

func TestBootstrapModelsFileMissingCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeModels(t, dir, testMannequinManifest, map[string][]byte{
		"LICENSE-quaternius.txt": readPackageModel(t, "LICENSE-quaternius.txt"),
	})
	if err := bootstrapModels(db); err == nil {
		t.Fatal("bootstrapModels must fail when the GLB file is missing")
	}
}

func TestBootstrapModelsMissingJointCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeModels(t, dir, `{"models":[{"id":"stub","file":"stub.glb"}]}`, map[string][]byte{
		"stub.glb": buildTestGLB(t, []string{"pelvis", "spine_01"}),
	})
	if err := bootstrapModels(db); err == nil {
		t.Fatal("bootstrapModels must fail on a model missing a required joint")
	}
}

func TestBootstrapModelsBadGLBCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeModels(t, dir, `{"models":[{"id":"stub","file":"stub.glb"}]}`, map[string][]byte{
		"stub.glb": []byte("not a glb"),
	})
	if err := bootstrapModels(db); err == nil {
		t.Fatal("bootstrapModels must fail on a malformed GLB")
	}
}

func TestBootstrapModelsStoreErrorCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeModels(t, dir, testMannequinManifest, map[string][]byte{
		"quaternius-mannequin.glb": readPackageModel(t, "quaternius-mannequin.glb"),
		"LICENSE-quaternius.txt":   readPackageModel(t, "LICENSE-quaternius.txt"),
	})
	server.failOn = "INSERT INTO"
	if err := bootstrapModels(db); err == nil {
		t.Fatal("bootstrapModels must fail when inserts fail")
	}
}

func TestGLBNodeNamesCoverage(t *testing.T) {
	names, err := glbNodeNames(buildTestGLB(t, []string{"pelvis", "Head"}))
	if err != nil {
		t.Fatalf("glbNodeNames: %v", err)
	}
	if len(names) != 2 || names[0] != "pelvis" || names[1] != "Head" {
		t.Fatalf("glbNodeNames = %v", names)
	}
	for name, data := range map[string][]byte{
		"short":       []byte("glTF"),
		"bad magic":   []byte("XXXX00000000000000000000"),
		"bad json":    append(append([]byte("glTF"), 2, 0, 0, 0, 24, 0, 0, 0, 4, 0, 0, 0, 0x4A, 0x53, 0x4F, 0x4E), []byte("{bad")...),
		"overrun":     append(append([]byte("glTF"), 2, 0, 0, 0, 24, 0, 0, 0, 255, 255, 0, 0, 0x4A, 0x53, 0x4F, 0x4E), []byte("{}")...),
		"wrong chunk": append(append([]byte("glTF"), 2, 0, 0, 0, 24, 0, 0, 0, 2, 0, 0, 0, 0x4E, 0x49, 0x42, 0x00), []byte("{}")...),
	} {
		if _, err := glbNodeNames(data); err == nil {
			t.Fatalf("glbNodeNames must fail on %s", name)
		}
	}
}

func TestBootstrapStillsCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeStills(t, dir, map[string]string{
		"barbell-full-squat.svg": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200"></svg>`,
		"notes.txt":              "not a still",
	})
	if err := os.MkdirAll(filepath.Join(dir, "stills", "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := bootstrapStills(db); err != nil {
		t.Fatalf("bootstrapStills: %v", err)
	}
	if !server.saw(`insert into "exercise_stills"`) {
		t.Fatal("no still inserts recorded")
	}
}

func TestBootstrapStillsMissingDirCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	chdirFeatures(t, map[string]string{})
	if err := bootstrapStills(db); err != nil {
		t.Fatalf("bootstrapStills without directory: %v", err)
	}
	if server.saw(`insert into "exercise_stills"`) {
		t.Fatal("still inserts recorded without a directory")
	}
}

func TestBootstrapStillsBadXMLCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeStills(t, dir, map[string]string{"burpee.svg": "not xml at all"})
	if err := bootstrapStills(db); err == nil {
		t.Fatal("bootstrapStills must fail on a malformed SVG")
	}
}

func TestBootstrapStillsWrongRootCoverage(t *testing.T) {
	db, _ := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeStills(t, dir, map[string]string{"burpee.svg": "<html></html>"})
	if err := bootstrapStills(db); err == nil {
		t.Fatal("bootstrapStills must fail on a non-SVG root element")
	}
}

func TestBootstrapStillsStoreErrorCoverage(t *testing.T) {
	db, server := openKnowledgeDB(t)
	dir := chdirFeatures(t, map[string]string{})
	writeStills(t, dir, map[string]string{
		"burpee.svg": `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200"></svg>`,
	})
	server.failOn = "INSERT INTO"
	if err := bootstrapStills(db); err == nil {
		t.Fatal("bootstrapStills must fail when inserts fail")
	}
}
