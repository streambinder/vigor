package llm

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/dm"
	"github.com/streambinder/vigor/llm/pipeline"
	"github.com/streambinder/vigor/model"
	"gorm.io/datatypes"
)

// fakeDecisionClient is a dm.Client returning canned answers, so the
// decision nodes run deterministically in unit tests. It lives here and
// nowhere else: production code never sees it.
type fakeDecisionClient struct {
	answers    map[string]dm.Answer
	noul       float64 // default probability for noul questions without a canned answer
	calls      [][]dm.Question
	lastStates []string
}

func (f *fakeDecisionClient) Decide(_ context.Context, state string, questions []dm.Question) (*dm.Result, error) {
	f.calls = append(f.calls, questions)
	f.lastStates = append(f.lastStates, state)
	out := make(map[string]dm.Answer, len(questions))
	for _, q := range questions {
		if a, ok := f.answers[q.ID]; ok {
			a.Kind = q.Kind
			out[q.ID] = a
			continue
		}
		switch q.Kind {
		case dm.KindNoul:
			out[q.ID] = dm.Answer{Kind: q.Kind, Noul: f.noul}
		case dm.KindScore:
			out[q.ID] = dm.Answer{Kind: q.Kind, Score: 0}
		case dm.KindChoice:
			first := ""
			for id := range q.Options {
				if first == "" || id < first {
					first = id
				}
			}
			out[q.ID] = dm.Answer{Kind: q.Kind, Choice: first}
		}
	}
	return &dm.Result{
		Model:   "typesafe/jev-1.13",
		Answers: out,
		Usage:   dm.Usage{InputTokens: 42, Cost: 0.001},
	}, nil
}

func installFakeDecisionClient(t *testing.T, fake *fakeDecisionClient) {
	t.Helper()
	prev := decisionClient
	decisionClient = fake
	t.Cleanup(func() { decisionClient = prev })
}

func TestDecideChunksAndMerges(t *testing.T) {
	fake := &fakeDecisionClient{noul: 0.9}
	installFakeDecisionClient(t, fake)

	questions := make([]dm.Question, 0, 40)
	for i := range 40 {
		questions = append(questions, dm.Question{
			ID:           "q" + strings.Repeat("x", i+1),
			Kind:         dm.KindNoul,
			Instructions: "claim",
		})
	}

	answers, step, err := decide("some state", questions)
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if len(answers) != 40 {
		t.Errorf("answers = %d, want 40", len(answers))
	}
	if len(fake.calls) != 2 {
		t.Errorf("client calls = %d, want 2 chunks", len(fake.calls))
	}
	if step.Kind != model.StepKindDM {
		t.Errorf("step kind = %q, want dm", step.Kind)
	}
	payload := step.DM.Data()
	if payload.Model != "typesafe/jev-1.13" || payload.Usage.InputTokens != 84 || len(payload.Questions) != 40 || len(payload.Answers) != 40 {
		t.Errorf("payload = model %q tokens %d questions %d answers %d, want merged step",
			payload.Model, payload.Usage.InputTokens, len(payload.Questions), len(payload.Answers))
	}
	if payload.StateHash == "" || payload.State != "some state" {
		t.Errorf("payload state/hash missing: %+v", payload)
	}
}

func TestDecideNoQuestionsSkipsClient(t *testing.T) {
	fake := &fakeDecisionClient{}
	installFakeDecisionClient(t, fake)

	answers, step, err := decide("state", nil)
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if len(answers) != 0 || len(fake.calls) != 0 {
		t.Errorf("answers %d calls %d, want none", len(answers), len(fake.calls))
	}
	if step.DM.Data().State != "state" {
		t.Errorf("step should still record the state, got %+v", step.DM.Data())
	}
}

func TestModifierAt(t *testing.T) {
	table := []float64{1.0, 0.9, 0.8}
	if got := modifierAt(table, -1); got != 1.0 {
		t.Errorf("below range = %v, want 1.0", got)
	}
	if got := modifierAt(table, 5); got != 0.8 {
		t.Errorf("above range = %v, want 0.8", got)
	}
	if got := modifierAt(table, 0.5); got < 0.949 || got > 0.951 {
		t.Errorf("midpoint = %v, want ~0.95", got)
	}
}

func TestHealthNodeShortCircuit(t *testing.T) {
	result, step, err := runHealthNode(nil)
	if err != nil {
		t.Fatalf("runHealthNode: %v", err)
	}
	if result.VolumeModifier != 1.0 || result.IntensityModifier != 1.0 {
		t.Errorf("modifiers = %v/%v, want 1.0/1.0", result.VolumeModifier, result.IntensityModifier)
	}
	if step.Kind != model.StepKindDM {
		t.Errorf("step kind = %q, want dm", step.Kind)
	}
}

func TestHealthNodeDecision(t *testing.T) {
	fake := &fakeDecisionClient{answers: map[string]dm.Answer{
		"recovery":      {Score: 3.0},
		"extend_warmup": {Noul: 0.9},
	}}
	installFakeDecisionClient(t, fake)

	snapshot := &model.HealthSnapshot{
		SleepHours: 5.5, SleepBaseline: 7.5, SleepDeviation: -26, SleepPresent: true,
		BaselineDays: 14,
	}
	result, step, err := runHealthNode(snapshot)
	if err != nil {
		t.Fatalf("runHealthNode: %v", err)
	}
	if result.VolumeModifier != 0.7 || result.IntensityModifier != 0.8 {
		t.Errorf("modifiers at score 3 = %v/%v, want 0.7/0.8", result.VolumeModifier, result.IntensityModifier)
	}
	if !result.ExtendWarmup {
		t.Error("extend warmup = false, want true at noul 0.9")
	}
	if !strings.Contains(result.Rationale, "significantly fatigued") {
		t.Errorf("rationale = %q, want level label", result.Rationale)
	}
	if step.DM.Data().Model != "typesafe/jev-1.13" {
		t.Errorf("step model = %q, want the fake's model", step.DM.Data().Model)
	}
}

func TestConstraintsNodeShortCircuit(t *testing.T) {
	result, _, err := runConstraintsNode([]model.Profile{{FirstName: "A"}})
	if err != nil {
		t.Fatalf("runConstraintsNode: %v", err)
	}
	if len(result.ContraindicatedPatterns) != 0 || len(result.Accommodations) != 0 {
		t.Errorf("result = %+v, want empty", result)
	}
}

func TestConstraintsNodeDecision(t *testing.T) {
	fake := &fakeDecisionClient{answers: map[string]dm.Answer{
		"pattern:deep-knee-flexion":        {Noul: 0.92},
		"pattern:overhead-pressing":        {Noul: 0.2},
		"accommodation:reduce-squat-depth": {Noul: 0.8},
	}, noul: 0.1}
	installFakeDecisionClient(t, fake)

	profile := model.Profile{
		Data: datatypes.JSON(`{"injuries":[{"description":"torn ACL","year":2023}],"limitations":["knee instability"]}`),
	}
	result, _, err := runConstraintsNode([]model.Profile{profile})
	if err != nil {
		t.Fatalf("runConstraintsNode: %v", err)
	}
	if len(result.ContraindicatedPatterns) != 1 || result.ContraindicatedPatterns[0] != "deep knee flexion" {
		t.Errorf("patterns = %v, want [deep knee flexion]", result.ContraindicatedPatterns)
	}
	if len(result.Accommodations) != 1 || result.Accommodations[0] != "reduce squat depth" {
		t.Errorf("accommodations = %v, want [reduce squat depth]", result.Accommodations)
	}
	if result.Summary == "no movement restrictions" || result.Summary == "" {
		t.Errorf("summary = %q, want a restriction summary", result.Summary)
	}
}

func TestMatchMovements(t *testing.T) {
	matched := matchMovements(
		"3 rounds: 10 pull-up, 15 push-up, then squat jumps",
		nil,
		[]MovementCandidate{
			{Name: "Pull Up"},
			{Name: "Push Up"},
			{Name: "Incline Push Up"},
			{Name: "Squat"},
			{Name: "Squat Jump"},
			{Name: "Deadlift"},
		},
	)
	want := map[string]bool{"Pull Up": true, "Push Up": true, "Squat Jump": true}
	if len(matched) != len(want) {
		t.Fatalf("matched = %v, want keys of %v", matched, want)
	}
	for _, m := range matched {
		if !want[m] {
			t.Errorf("unexpected match %q", m)
		}
	}
}

// an Italian article names its movements in Italian: the catalog aliases
// must still pin them, returning canonical names (regression: a Spider-Man
// program article pinned only the two movements it spelled in English,
// and the delivered session silently lost pull-ups, dips and sit-ups).
func TestMatchMovementsItalianAliases(t *testing.T) {
	article := "Il Ladder prevede 1 trazione alla sbarra, 2 dip su parallele, " +
		"3 push-up, 4 addominali e 5 air squat, salendo fino a 10-20-30-40-50."
	matched := matchMovements(
		"replica questo allenamento",
		[]string{article},
		[]MovementCandidate{
			{Name: "Pull-Up", Aliases: []string{"trazioni", "trazioni alla sbarra"}},
			{Name: "Chest Dip", Aliases: []string{"dip", "parallele", "petto dip"}},
			{Name: "3/4 Sit-Up", Aliases: []string{"addominali"}},
			{Name: "Push-Up", Aliases: []string{"piegamenti", "flessioni"}},
			{Name: "Air Squat", Aliases: []string{"squat a corpo libero"}},
			{Name: "Deadlift", Aliases: []string{"stacco da terra"}},
		},
	)
	want := map[string]bool{
		"Pull-Up": true, "Chest Dip": true, "3/4 Sit-Up": true,
		"Push-Up": true, "Air Squat": true,
	}
	if len(matched) != len(want) {
		t.Fatalf("matched = %v, want keys of %v", matched, want)
	}
	for _, m := range matched {
		if !want[m] {
			t.Errorf("unexpected match %q", m)
		}
	}
}

// a long article names far more than it programs: scaling asides,
// examples, and a title that collides with catalog names. Only the
// movements the programs actually prescribe — contiguous, recurring
// or inside a rep scheme — may pin; scattered co-occurrence and
// one-off mentions must not (regression: a Spider-Man program article
// pinned spider curls and L-sit variants matched from transliterated
// aliases and stray words, and the dip fell out of the pin cap).
func TestMatchMovementsNoisyArticle(t *testing.T) {
	article := "L'allenamento a corpo libero di Spider-Man. " +
		"Primo programma: timer di 20 minuti, più round possibili di 5 trazioni, " +
		"10 piegamenti e 15 squat a corpo libero; in inglese 5 pull up, 10 push up " +
		"e 15 air squat. Varianti facilitate: trazioni orizzontali al TRX, " +
		"piegamenti sulle ginocchia, squat box con una panca dietro. " +
		"Secondo programma: 1 trazione alla sbarra, 2 dip, 3 push-up, 4 addominali " +
		"e 5 squat; si sale fino a 10 trazioni, 20 dip, 30 push-up, 40 addominali " +
		"e 50 squat, poi si ridiscende fino a 1. Esecuzione: 1 trazione alla " +
		"sbarra, 2 dip, 3 piegamenti, 4 sit-up, 5 squat. " +
		"Scalare: dip facilitati su panca o box, crunch a terra, box squat. " +
		"Personalizzare: elastici per le trazioni, esercizi monoarto come " +
		"one arm push up e one arm pull up, manubri tra le gambe come zavorra."
	matched := matchMovements(
		"replica l'allenamento di questo articolo",
		[]string{article},
		[]MovementCandidate{
			{Name: "Pull-Up", Aliases: []string{"trazioni", "trazioni alla sbarra"}},
			{Name: "Chest Dip", Aliases: []string{"dip", "parallele", "petto dip"}},
			{Name: "3/4 Sit-Up", Aliases: []string{"addominali"}},
			{Name: "Push-Up", Aliases: []string{"piegamenti", "flessioni"}},
			{Name: "Air Squat", Aliases: []string{"squat a corpo libero"}},
			{Name: "Cable Spider Curl", Aliases: []string{"cavo spider curl", "трос spider сгибание"}},
			{Name: "Dumbbell Spider Curl", Aliases: []string{"manubrio spider curl", "гантель spider сгибание"}},
			{Name: "L-Pull-Up", Aliases: []string{"l trazione su", "l тяга вверх"}},
			{Name: "L-Sit", Aliases: []string{"l seduta", "l сидя"}},
			{Name: "Ring L-Sit", Aliases: []string{"anello l seduta", "кольцо l сидя"}},
			{Name: "Spider Crawl Push Up", Aliases: []string{"spider strisciata spinta su", "spider ползание жим вверх"}},
			{Name: "One Arm Push Up", Aliases: []string{"spinta su un braccio"}},
			{Name: "Crunch Floor", Aliases: []string{"crunch a terra"}},
		},
	)
	want := map[string]bool{
		"Pull-Up": true, "Chest Dip": true, "3/4 Sit-Up": true,
		"Push-Up": true, "Air Squat": true,
	}
	if len(matched) != len(want) {
		t.Fatalf("matched = %v, want keys of %v", matched, want)
	}
	for _, m := range matched {
		if !want[m] {
			t.Errorf("unexpected match %q", m)
		}
	}
}

func TestDeriveNodeDecision(t *testing.T) {
	fake := &fakeDecisionClient{answers: map[string]dm.Answer{
		"methodology":        {Choice: "circuit"},
		"goal:hypertrophy":   {Noul: 0.9},
		"muscle:chest":       {Noul: 0.8},
		"equipment:dumbbell": {Noul: 0.7},
		"explicit_program":   {Noul: 0.95},
	}, noul: 0.1}
	installFakeDecisionClient(t, fake)

	derived, step, err := DeriveFreeTextParams(DeriveRequest{
		FreeText:           "allenamento a circuito con 10 pull-up e 20 push-up",
		Methodologies:      []model.Methodology{{ID: "circuit", Name: "Circuit", Description: "rounds"}, {ID: "strength", Name: "Strength", Description: "heavy"}},
		AllGoals:           []model.Goal{{ID: "hypertrophy", Description: "muscle mass"}, {ID: "endurance", Description: "stamina"}},
		ValidMuscles:       []string{"chest", "legs"},
		ValidEquipment:     []string{"dumbbell", "barbell"},
		MovementCandidates: []MovementCandidate{{Name: "Pull Up"}, {Name: "Push Up"}, {Name: "Squat"}},
	})
	if err != nil {
		t.Fatalf("DeriveFreeTextParams: %v", err)
	}
	if derived.Methodology != "circuit" {
		t.Errorf("methodology = %q, want circuit", derived.Methodology)
	}
	if len(derived.Goals) != 1 || derived.Goals[0] != "hypertrophy" {
		t.Errorf("goals = %v, want [hypertrophy]", derived.Goals)
	}
	if len(derived.Muscles) != 1 || derived.Muscles[0] != "chest" {
		t.Errorf("muscles = %v, want [chest]", derived.Muscles)
	}
	if len(derived.Equipment) != 1 || derived.Equipment[0] != "dumbbell" {
		t.Errorf("equipment = %v, want [dumbbell]", derived.Equipment)
	}
	if !derived.ExplicitProgram || len(derived.Movements) != 2 {
		t.Errorf("explicit = %v movements = %v, want explicit with 2 movements", derived.ExplicitProgram, derived.Movements)
	}
	if !strings.Contains(derived.Summary, "Methodology: circuit") || !strings.Contains(derived.Summary, "pull-up") {
		t.Errorf("summary = %q, want derivation + request words", derived.Summary)
	}
	if step.Kind != model.StepKindDM {
		t.Errorf("step kind = %q, want dm", step.Kind)
	}
}

func TestMuscleTargetingNodeDecision(t *testing.T) {
	fake := &fakeDecisionClient{answers: map[string]dm.Answer{
		"muscle:back":  {Score: 1.9},
		"muscle:chest": {Score: 1.0},
		"muscle:legs":  {Score: 0.1},
	}}
	installFakeDecisionClient(t, fake)

	result, _, err := runMuscleTargetingNode(
		nil, nil,
		map[string]int{"back": 12, "chest": 10, "legs": 14},
		pipeline.ConstraintExtraction{}, pipeline.HealthAssessment{VolumeModifier: 1, IntensityModifier: 1},
		pipeline.HistoryAnalysis{}, "", false,
	)
	if err != nil {
		t.Fatalf("runMuscleTargetingNode: %v", err)
	}
	if len(result.PrimaryMuscles) != 1 || result.PrimaryMuscles[0] != "back" {
		t.Errorf("primary = %v, want [back]", result.PrimaryMuscles)
	}
	if len(result.SecondaryMuscles) != 1 || result.SecondaryMuscles[0] != "chest" {
		t.Errorf("secondary = %v, want [chest]", result.SecondaryMuscles)
	}
	if len(result.AvoidMuscles) != 1 || result.AvoidMuscles[0] != "legs" {
		t.Errorf("avoid = %v, want [legs]", result.AvoidMuscles)
	}
}

func TestHistoryNodeProgressions(t *testing.T) {
	fake := &fakeDecisionClient{answers: map[string]dm.Answer{
		"progression:bench-press": {Choice: "increase_weight"},
		"progression:deadlift":    {Choice: "maintain"},
	}}
	installFakeDecisionClient(t, fake)

	trainingID := uuid.New()
	trainings := []model.Training{
		{
			Name: "Push Day",
			Routines: []model.Routine{{Blocks: []model.Block{{Activities: []model.Activity{
				{ExerciseID: "bench-press", WeightKg: 60, Reps: 8},
				{ExerciseID: "squat", WeightKg: 80, Reps: 5},
				{ExerciseID: "deadlift", WeightKg: 100, Reps: 5},
			}}}}},
		},
	}
	feedback := map[uuid.UUID]model.TrainingFeedback{
		trainingID: {},
	}
	// the training carries no id of its own in this fixture; key feedback by its zero id instead
	trainings[0].ID = trainingID
	feedback[trainingID] = model.TrainingFeedback{
		ActivityFeedback: datatypes.JSON(`{"bench-press":"too_easy","squat":"impossible","deadlift":"too_hard"}`),
	}

	result, step, err := runHistoryNode(trainings, feedback, nil)
	if err != nil {
		t.Fatalf("runHistoryNode: %v", err)
	}
	if len(result.AvoidExercises) != 1 || result.AvoidExercises[0] != "squat" {
		t.Errorf("avoid = %v, want [squat]", result.AvoidExercises)
	}
	byExercise := map[string]pipeline.ProgressionSignal{}
	for _, p := range result.Progressions {
		byExercise[p.ExerciseID] = p
	}
	bench, ok := byExercise["bench-press"]
	if !ok || bench.Action != "increase_weight" || float64(bench.ToWeight) != 63 {
		t.Errorf("bench progression = %+v, want increase_weight to 63kg", bench)
	}
	squat, ok := byExercise["squat"]
	if !ok || squat.Action != "replace" {
		t.Errorf("squat progression = %+v, want replace", squat)
	}
	if _, ok := byExercise["deadlift"]; ok {
		t.Errorf("maintain must not produce a progression, got %+v", byExercise["deadlift"])
	}
	if step.Kind != model.StepKindDM {
		t.Errorf("step kind = %q, want dm", step.Kind)
	}
	if len(result.RecentNames) != 1 || result.RecentNames[0] != "Push Day" {
		t.Errorf("recent names = %v", result.RecentNames)
	}
}

func TestHistoryNodeEmpty(t *testing.T) {
	result, step, err := runHistoryNode(nil, nil, nil)
	if err != nil {
		t.Fatalf("runHistoryNode: %v", err)
	}
	if len(result.Progressions) != 0 || step.Kind != model.StepKindDM {
		t.Errorf("result = %+v step kind %q, want empty dm step", result, step.Kind)
	}
}

// twoProgramsArticle carries two distinct programs under their own
// headings plus prose sections that program nothing. It is a distilled
// paraphrase written for this test, not copied article text.
const twoProgramsArticle = `Allenamento a corpo libero completo
Due schemi diversi per allenarsi senza attrezzi, da scegliere in base
al tempo e al livello. Entrambi lavorano su spinta, trazione, gambe
e addome con movimenti base del calisthenics moderno.

Programma A: il Circuito Veloce
Timer di 20 minuti, più round possibili di 5 trazioni alla sbarra,
10 push-up e 15 air squat, senza pause lunghe tra un round e l'altro.
Il circuito veloce chiede densità: 5 trazioni alla sbarra, 10 push-up
e 15 air squat a ogni giro, restando vicino alla sbarra per non
perdere tempo negli spostamenti durante tutta la seduta a corpo
libero, in palestra oppure a casa propria con un minimo di spazio.

Programma B: la Scala
Si parte da 1 trazione alla sbarra, 2 dip, 3 push-up, 4 addominali
e 5 squat, poi si sale fino a 10 trazioni alla sbarra, 20 dip,
30 push-up, 40 addominali e 50 squat, e infine si ridiscende fino
a 1 trazione alla sbarra, 2 dip, 3 push-up, 4 addominali e 5 squat.
La scala completa dura circa un'ora, con recuperi liberi a
sensazione tra un round e l'altro in base al proprio livello.

Consigli pratici
Dormire bene, mangiare proteine a sufficienza e bere acqua aiuta
il recupero tra una seduta e l'altra, qualunque schema si scelga
per allenarsi con costanza durante la settimana lavorativa intera.`

func programTestCandidates() []MovementCandidate {
	return []MovementCandidate{
		{Name: "Pull-Up", Aliases: []string{"trazioni", "trazioni alla sbarra"}},
		{Name: "Chest Dip", Aliases: []string{"dip", "parallele", "petto dip"}},
		{Name: "3/4 Sit-Up", Aliases: []string{"addominali"}},
		{Name: "Push-Up", Aliases: []string{"piegamenti", "flessioni"}},
		{Name: "Air Squat", Aliases: []string{"squat a corpo libero"}},
	}
}

func movementSet(movements []string) map[string]bool {
	set := make(map[string]bool, len(movements))
	for _, m := range movements {
		set[m] = true
	}
	return set
}

func TestSegmentPrograms(t *testing.T) {
	programs := segmentPrograms([]string{twoProgramsArticle}, programTestCandidates())
	if len(programs) != 2 {
		t.Fatalf("programs = %d, want 2", len(programs))
	}
	circuit := movementSet(programs[0].Movements)
	for _, want := range []string{"Pull-Up", "Push-Up", "Air Squat"} {
		if !circuit[want] {
			t.Errorf("circuit movements = %v, missing %q", programs[0].Movements, want)
		}
	}
	if circuit["Chest Dip"] || circuit["3/4 Sit-Up"] {
		t.Errorf("circuit movements = %v, must not blend the ladder in", programs[0].Movements)
	}
	ladder := movementSet(programs[1].Movements)
	for _, want := range []string{"Pull-Up", "Chest Dip", "Push-Up", "3/4 Sit-Up", "Air Squat"} {
		if !ladder[want] {
			t.Errorf("ladder movements = %v, missing %q", programs[1].Movements, want)
		}
	}
	if !strings.Contains(programs[1].Title, "Scala") {
		t.Errorf("ladder title = %q, want the Scala heading", programs[1].Title)
	}
}

func TestSegmentProgramsSingleProgram(t *testing.T) {
	article := "Circuito semplice\n" +
		"Ripeti per 20 minuti 5 trazioni alla sbarra, 10 push-up e 15 air squat, " +
		"poi ancora 5 trazioni alla sbarra, 10 push-up e 15 air squat senza pause " +
		"lunghe, restando vicino alla sbarra per tutta la seduta a corpo libero, " +
		"in palestra oppure a casa propria con un minimo di spazio a disposizione."
	programs := segmentPrograms([]string{article}, programTestCandidates())
	if len(programs) != 1 {
		t.Fatalf("programs = %d, want 1", len(programs))
	}

	prose := "Consigli generali\n" +
		"Dormire bene e mangiare proteine aiuta il recupero tra una seduta " +
		"e l'altra, qualunque schema si scelga per allenarsi con costanza " +
		"durante la settimana, a casa oppure in palestra con gli amici."
	if got := segmentPrograms([]string{prose}, programTestCandidates()); len(got) != 0 {
		t.Fatalf("prose programs = %d, want 0", len(got))
	}
}

func TestDeriveProgramChoice(t *testing.T) {
	fake := &fakeDecisionClient{answers: map[string]dm.Answer{
		"methodology":      {Choice: "amrap"},
		"program":          {Choice: "program_1"},
		"explicit_program": {Noul: 0.95},
	}, noul: 0.1}
	installFakeDecisionClient(t, fake)

	derived, step, err := DeriveFreeTextParams(DeriveRequest{
		FreeText:           "voglio replicare l'allenamento di questo articolo",
		Articles:           []string{twoProgramsArticle},
		Methodologies:      []model.Methodology{{ID: "amrap", Name: "AMRAP", Description: "rounds against the clock"}},
		MovementCandidates: programTestCandidates(),
	})
	if err != nil {
		t.Fatalf("DeriveFreeTextParams: %v", err)
	}
	if !derived.ExplicitProgram {
		t.Error("want explicit program")
	}
	got := movementSet(derived.Movements)
	for _, want := range []string{"Pull-Up", "Chest Dip", "Push-Up", "3/4 Sit-Up", "Air Squat"} {
		if !got[want] {
			t.Errorf("movements = %v, missing ladder movement %q", derived.Movements, want)
		}
	}
	if !strings.Contains(derived.ProgramText, "la Scala") {
		t.Errorf("program text missing the chosen segment: %.80q", derived.ProgramText)
	}
	if strings.Contains(derived.Summary, "Circuito Veloce") {
		t.Errorf("summary blends the unchosen program: %.200q", derived.Summary)
	}
	asked := false
	for _, call := range fake.calls {
		for _, q := range call {
			if q.ID == "program" {
				asked = true
			}
		}
	}
	if !asked {
		t.Error("program choice question was not asked")
	}
	if step.Kind != model.StepKindDM {
		t.Errorf("step kind = %q, want dm", step.Kind)
	}
}
