package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func TestProfileAccessors(t *testing.T) {
	birth := time.Now().AddDate(-30, 0, 0)
	p := &Profile{Birthdate: birth}
	if age := p.Age(); age < 29 || age > 30 {
		t.Errorf("Age = %d", age)
	}
	// birthday later in the year than today: age decrements.
	future := time.Now().AddDate(-20, 6, 0)
	if age := (&Profile{Birthdate: future}).Age(); age < 18 || age > 20 {
		t.Errorf("Age future = %d", age)
	}

	valid := &Profile{Data: datatypes.JSON(`{"goals":["strength"],"injuries":[{"description":"knee","year":2020}],"limitations":["overhead"],"conditions":["asthma"],"preferences":{"exercises":["push-up"],"equipment":["band"]}}`)}
	if got := valid.Goals(); len(got) != 1 || got[0] != "strength" {
		t.Errorf("Goals = %v", got)
	}
	if got := valid.Injuries(); len(got) != 1 || got[0].Description != "knee" {
		t.Errorf("Injuries = %v", got)
	}
	if got := valid.Limitations(); len(got) != 1 {
		t.Errorf("Limitations = %v", got)
	}
	if got := valid.Conditions(); len(got) != 1 {
		t.Errorf("Conditions = %v", got)
	}
	if got := valid.FavoriteExercises(); len(got) != 1 {
		t.Errorf("FavoriteExercises = %v", got)
	}
	if got := valid.FavoriteEquipment(); len(got) != 1 {
		t.Errorf("FavoriteEquipment = %v", got)
	}

	invalid := &Profile{Data: datatypes.JSON(`{oops`)}
	if invalid.Goals() != nil || invalid.Injuries() != nil || invalid.Limitations() != nil || invalid.Conditions() != nil || invalid.FavoriteExercises() != nil || invalid.FavoriteEquipment() != nil {
		t.Error("invalid profile data must yield nil accessors")
	}
	noPrefs := &Profile{Data: datatypes.JSON(`{"goals":[]}`)}
	if noPrefs.FavoriteExercises() != nil || noPrefs.FavoriteEquipment() != nil {
		t.Error("missing preferences must yield nil favorites")
	}
}

func TestMethodologyAccessors(t *testing.T) {
	m := &Methodology{}
	if err := m.SetWork(MethodologyWork{MinDifficulty: 1, MaxDifficulty: 5}); err != nil {
		t.Fatalf("SetWork: %v", err)
	}
	if got := m.GetWork(); got.MinDifficulty != 1 || got.MaxDifficulty != 5 {
		t.Errorf("GetWork = %+v", got)
	}
	if err := m.SetExercisesPerHour(ExerciseDensity{Min: 2, Max: 6}); err != nil {
		t.Fatalf("SetExercisesPerHour: %v", err)
	}
	if got := m.GetExercisesPerHour(); got.Min != 2 || got.Max != 6 {
		t.Errorf("GetExercisesPerHour = %+v", got)
	}
	bad := &Methodology{Work: datatypes.JSON(`{bad`), ExercisesPerHour: datatypes.JSON(`{bad`)}
	if got := bad.GetWork(); got != (MethodologyWork{}) {
		t.Errorf("bad GetWork = %+v", got)
	}
	if got := bad.GetExercisesPerHour(); got != (ExerciseDensity{Min: 4, Max: 8}) {
		t.Errorf("bad GetExercisesPerHour = %+v", got)
	}
	zero := &Methodology{ExercisesPerHour: datatypes.JSON(`{"min":0,"max":0}`)}
	if got := zero.GetExercisesPerHour(); got.Max != 8 {
		t.Errorf("zero density fallback = %+v", got)
	}
}

func TestFlowSessionPosesAndAfterFind(t *testing.T) {
	s := &FlowSession{}
	poses := []FlowPose{{ExerciseID: "pigeon", Name: "Pigeon", Duration: 30, Rest: 5}, {ExerciseID: "cobra", Duration: 20}}
	if err := s.SetPoses(poses); err != nil {
		t.Fatalf("SetPoses: %v", err)
	}
	got, err := s.GetPoses()
	if err != nil || len(got) != 2 || got[0].ExerciseID != "pigeon" {
		t.Fatalf("GetPoses = %v, %v", got, err)
	}
	bad := &FlowSession{Poses: []byte(`{bad`)}
	if _, err := bad.GetPoses(); err == nil {
		t.Error("invalid poses must error")
	}
	// AfterFind with no trajectory and with a trajectory carrying steps.
	if err := s.AfterFind(nil); err != nil {
		t.Errorf("AfterFind nil trajectory: %v", err)
	}
	s.Trajectory = &Trajectory{Steps: []ModelStep{NewLLMStep(LLMStep{Model: "m1"})}}
	if err := s.AfterFind(nil); err != nil {
		t.Errorf("AfterFind: %v", err)
	}
	if len(s.Trajectory.Models) != 1 || s.Trajectory.Models[0] != "m1" {
		t.Errorf("trajectory models = %v", s.Trajectory.Models)
	}
	if FlowSessionSchema.JSONSchema.Name != "flow_session_schema" {
		t.Errorf("schema = %+v", FlowSessionSchema)
	}
}

func TestHealthTablesAndSignal(t *testing.T) {
	if (HealthSleepDaily{}).TableName() != "health_sleep_daily" {
		t.Error("sleep table name")
	}
	if (HealthRecoveryDaily{}).TableName() != "health_recovery_daily" {
		t.Error("recovery table name")
	}
	if (&HealthSnapshot{}).HasRecoverySignal() {
		t.Error("empty snapshot must have no recovery signal")
	}
	for _, snap := range []*HealthSnapshot{{SleepPresent: true}, {HRVPresent: true}, {RHRPresent: true}} {
		if !snap.HasRecoverySignal() {
			t.Errorf("snapshot %+v must have a signal", snap)
		}
	}
}

func TestModelStepConstructorsAndNames(t *testing.T) {
	if (ModelStep{}).TableName() != "model_steps" {
		t.Error("model step table name")
	}
	if (Trajectory{}).TableName() != "trajectories" {
		t.Error("trajectory table name")
	}
	if (RefreshToken{}).TableName() != "tokens" {
		t.Error("token table name")
	}
	llmStep := NewLLMStep(LLMStep{Model: "llm-model"})
	if llmStep.Kind != StepKindLLM || llmStep.ModelName() != "llm-model" {
		t.Errorf("llm step = %+v name %q", llmStep, llmStep.ModelName())
	}
	dmStep := NewDMStep(DMStep{Model: "dm-model"})
	if dmStep.Kind != StepKindDM || dmStep.ModelName() != "dm-model" {
		t.Errorf("dm step = %+v name %q", dmStep, dmStep.ModelName())
	}
	if (ModelStep{Kind: "other"}).ModelName() != "" {
		t.Error("unknown step kind must have an empty model name")
	}
}

func TestTrainingErrorDaysActivitiesClone(t *testing.T) {
	verr := &ValidationError{Code: "empty_name", Message: "training name is empty"}
	if verr.Error() != "training name is empty" {
		t.Errorf("Error = %q", verr.Error())
	}
	if verr.Reason() != "validation_error:empty_name" {
		t.Errorf("Reason = %q", verr.Reason())
	}

	completed := time.Now().AddDate(0, 0, -3)
	tr := Training{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		CompletedAt: &completed,
		CreatedAt:   time.Now().AddDate(0, 0, -10),
		Routines: []Routine{
			{Type: "warmup", Position: 1, Blocks: []Block{{Position: 1, Activities: []Activity{{ExerciseID: "arm-circles", Position: 1}}}}},
			{Type: "work", Position: 2, Blocks: []Block{{Position: 1, Activities: []Activity{
				{ExerciseID: "push-up", Position: 1},
				{ExerciseID: "push-up", Position: 2},
				{ExerciseID: "air-squat", Position: 3},
			}}}},
		},
	}
	if err := tr.AfterFind(nil); err != nil {
		t.Errorf("AfterFind: %v", err)
	}
	if days := tr.DaysSince(); days < 2 || days > 3 {
		t.Errorf("DaysSince completed = %d", days)
	}
	uncompleted := Training{CreatedAt: time.Now().AddDate(0, 0, -5)}
	if days := uncompleted.DaysSince(); days < 4 || days > 5 {
		t.Errorf("DaysSince created = %d", days)
	}
	acts := tr.Activities()
	if len(acts) != 2 {
		t.Fatalf("Activities = %d, want unique work activities only", len(acts))
	}

	newUser := uuid.New()
	clone := tr.Clone(newUser)
	if clone.ID != (uuid.UUID{}) || clone.UserID != newUser || clone.ParentID == nil || *clone.ParentID != tr.ID {
		t.Errorf("clone identity = %+v", clone)
	}
	if clone.CompletedAt != nil || clone.Trajectory != nil || clone.GymID != nil {
		t.Error("clone must clear completion, trajectory and gym")
	}
	if len(clone.Routines) != 2 || clone.Routines[1].Blocks[0].Activities[0].ID != "" || clone.Routines[1].Position != 2 {
		t.Errorf("clone routines = %+v", clone.Routines)
	}
	// the original stays intact.
	if tr.Routines[0].ID == "cleared-by-clone" {
		t.Error("clone mutated the original")
	}
}
