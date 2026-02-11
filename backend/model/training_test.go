package model

import "testing"

// helper to build a minimal training with given methodology and routines
func makeTraining(methodology string, routines []Routine) Training {
	return Training{Methodology: methodology, Routines: routines}
}

func routine(typ string, rest int, blocks []Block) Routine {
	return Routine{Type: typ, Rest: rest, Blocks: blocks}
}

func block(repeats, rest int, activities []Activity) Block {
	return Block{Repeats: repeats, Rest: rest, Activities: activities}
}

func act(duration, reps, rest int) Activity {
	return Activity{Duration: duration, Reps: reps, Rest: rest}
}

func actWeighted(exerciseID string, weightKg int, modifiers []string) Activity {
	return Activity{ExerciseID: exerciseID, Reps: 10, WeightKg: weightKg, Modifiers: modifiers}
}

func TestActivityWorkDuration(t *testing.T) {
	tests := []struct {
		name     string
		activity Activity
		expected int
	}{
		{"explicit duration", act(45, 0, 0), 45},
		{"reps high", act(0, 10, 0), 40},                   // 10*4=40
		{"reps low", act(0, 5, 0), 20},                     // 5*4=20
		{"zero reps zero duration", act(0, 0, 0), 30},      // fallback
		{"duration takes precedence", act(60, 12, 0), 60},  // duration > 0 wins
	}
	for _, tt := range tests {
		if got := activityWorkDuration(tt.activity); got != tt.expected {
			t.Errorf("%s: activityWorkDuration() = %d, want %d", tt.name, got, tt.expected)
		}
	}
}

func TestCalculateDuration_Strength(t *testing.T) {
	// 2 blocks, 3 repeats each, 2 activities per block
	// block rest=60s between repeats, activity rest=30s between activities
	tr := makeTraining("strength", []Routine{
		routine("warmup", 0, []Block{
			block(1, 0, []Activity{act(60, 0, 0)}),
		}),
		routine("work", 0, []Block{
			block(3, 60, []Activity{
				act(0, 10, 30), // 40s work
				act(0, 10, 0),  // 40s work
			}),
			block(3, 60, []Activity{
				act(0, 10, 30), // 40s work
				act(0, 10, 0),  // 40s work
			}),
		}),
		routine("cooldown", 0, []Block{
			block(1, 0, []Activity{act(60, 0, 0)}),
		}),
	})

	got := tr.CalculateDuration()

	// warmup: 60s (not last routine → no rest suppression, but block rest=0)
	// work block 1 (3 repeats): each repeat = 40 + 30(act) + 40 + 60(block) = 170, all 3 = 510
	//   (work is routine index 1 of 3, not last routine, so no block rest suppression)
	// work block 2 (3 repeats): same structure, also not last routine → 170*3 = 510
	// cooldown: 60s (last routine, last block, last repeat, last activity → no rest)
	// total = 60 + 510 + 510 + 60 = 1140

	if got != 1140 {
		t.Errorf("strength CalculateDuration() = %d, want 1140", got)
	}
}

func TestCalculateDuration_RestMerging(t *testing.T) {
	// activity rest 30s followed by block rest 90s → keep 90s (merging)
	tr := makeTraining("strength", []Routine{
		routine("work", 0, []Block{
			block(2, 90, []Activity{
				act(0, 10, 30), // 40s work + 30s rest (not last in repeat)
				act(0, 10, 0),  // 40s work (last in repeat → block rest)
			}),
			block(1, 0, []Activity{
				act(0, 10, 0), // 40s work (last activity of training)
			}),
		}),
	})

	got := tr.CalculateDuration()

	// block 1:
	//   repeat 1: 40 + 30(act rest) + 40 + 90(block rest) = 200
	//   repeat 2: 40 + 30(act rest) + 40 + 90(block rest) = 200
	//     but repeat 2 is last repeat of block 1, block 1 is not last block → block rest added
	//     however this block rest (90) follows block 1 repeat 2, and then...
	//     actually: isLastOfTraining = isLastRoutine(yes) && isLastBlock(no) → false, so block rest is added
	// block 2:
	//   repeat 1: 40 (last activity of training, no rest)
	// but: block rest 90 from block1 repeat2 is followed by work 40 from block2, no consecutive rests

	// let me recalculate:
	// repeat 1: work(40) + actRest(30) + work(40) + blockRest(90) = 200
	// repeat 2: work(40) + actRest(30) + work(40) + blockRest(90) = 200
	// block 2: work(40)
	// total = 200 + 200 + 40 = 440

	if got != 440 {
		t.Errorf("rest merging CalculateDuration() = %d, want 440", got)
	}
}

func TestCalculateDuration_RestMerging_ConsecutiveRests(t *testing.T) {
	// test actual rest merging: routine rest after block rest → keep longer
	tr := makeTraining("circuit", []Routine{
		routine("work", 120, []Block{
			block(1, 60, []Activity{
				act(30, 0, 0),
			}),
		}),
		routine("work", 0, []Block{
			block(1, 0, []Activity{
				act(30, 0, 0),
			}),
		}),
	})

	got := tr.CalculateDuration()

	// routine 1:
	//   block 1, repeat 1: work(30), last activity of training? no (there's another routine)
	//     isLastInRepeat=true, block.Rest=60, isLastOfTraining? isLastRoutine=false → no
	//     → blockRest(60) added
	//   routine rest: 120, not last routine → addRest(120)
	//     consecutive rest: 60 < 120 → replace with 120
	// routine 2:
	//   block 1, repeat 1: work(30), last of training → no rest
	// total = 30 + 120 + 30 = 180

	if got != 180 {
		t.Errorf("consecutive rest merging CalculateDuration() = %d, want 180", got)
	}
}

func TestCalculateDuration_EMOM(t *testing.T) {
	tr := makeTraining("emom", []Routine{
		routine("warmup", 0, []Block{
			block(1, 0, []Activity{act(120, 0, 0)}),
		}),
		routine("work", 0, []Block{
			block(10, 0, []Activity{act(0, 5, 0)}), // 10 rounds × 60s = 600s
			block(5, 0, []Activity{act(0, 8, 0)}),  // 5 rounds × 60s = 300s
		}),
		routine("cooldown", 0, []Block{
			block(1, 0, []Activity{act(60, 0, 0)}),
		}),
	})

	got := tr.CalculateDuration()

	// warmup: 120s
	// work: 10*60 + 5*60 = 900s
	// cooldown: 60s
	// total = 1080s

	if got != 1080 {
		t.Errorf("emom CalculateDuration() = %d, want 1080", got)
	}
}

func TestCalculateDuration_ForTime(t *testing.T) {
	tr := makeTraining("for_time", []Routine{
		routine("warmup", 0, []Block{
			block(1, 0, []Activity{act(120, 0, 0)}),
		}),
		routine("work", 0, []Block{
			block(3, 0, []Activity{act(0, 20, 0)}),
		}),
		routine("cooldown", 0, []Block{
			block(1, 0, []Activity{act(60, 0, 0)}),
		}),
	})

	got := tr.CalculateDuration()

	// for_time uses interval estimation like strength:
	// warmup(120) + work(3 repeats × 80s = 240) + cooldown(60) = 420
	if got != 420 {
		t.Errorf("for_time CalculateDuration() = %d, want 420", got)
	}
}

func TestSetDuration_AMRAP(t *testing.T) {
	tr := makeTraining("amrap", []Routine{
		routine("warmup", 0, []Block{
			block(1, 0, []Activity{act(120, 0, 0)}),
		}),
		routine("work", 0, []Block{
			block(1, 0, []Activity{act(0, 10, 0)}),
		}),
		routine("cooldown", 0, []Block{
			block(1, 0, []Activity{act(60, 0, 0)}),
		}),
	})

	tr.SetDuration(30) // 30 minutes = 1800s

	// warmup = 120s, cooldown = 60s
	// work = 1800 - 120 - 60 = 1620s
	if tr.Duration != 1620 {
		t.Errorf("amrap SetDuration(30): Duration = %d, want 1620", tr.Duration)
	}

	// CalculateDuration should return total including warmup + work + cooldown
	total := tr.CalculateDuration()
	if total != 1800 {
		t.Errorf("amrap CalculateDuration() after SetDuration(30) = %d, want 1800", total)
	}
}

func TestSetDuration_AMRAP_MinimumWork(t *testing.T) {
	// if warmup+cooldown exceeds requested duration, work should be clamped to 60s
	tr := makeTraining("amrap", []Routine{
		routine("warmup", 0, []Block{
			block(1, 0, []Activity{act(600, 0, 0)}), // 10min warmup
		}),
		routine("work", 0, []Block{
			block(1, 0, []Activity{act(0, 10, 0)}),
		}),
		routine("cooldown", 0, []Block{
			block(1, 0, []Activity{act(300, 0, 0)}), // 5min cooldown
		}),
	})

	tr.SetDuration(10) // 10 minutes = 600s, but warmup+cooldown = 900s

	if tr.Duration != 60 {
		t.Errorf("amrap SetDuration minimum: Duration = %d, want 60", tr.Duration)
	}
}

func TestSetDuration_Strength(t *testing.T) {
	tr := makeTraining("strength", []Routine{
		routine("work", 0, []Block{
			block(3, 60, []Activity{
				act(0, 10, 30),
				act(0, 10, 0),
			}),
		}),
	})

	tr.SetDuration(45) // value doesn't matter for non-amrap

	// strength SetDuration → CalculateDuration()
	// repeat 1: 40 + 30 + 40 + 60 = 170
	// repeat 2: 40 + 30 + 40 + 60 = 170
	// repeat 3: 40 + 30 + 40 = 110 (last of training)
	// total = 450
	if tr.Duration != 450 {
		t.Errorf("strength SetDuration: Duration = %d, want 450", tr.Duration)
	}
}

func TestCalculateDuration_SkipWarmupCooldown(t *testing.T) {
	// no warmup/cooldown routines — verify all methodology paths handle this
	tests := []struct {
		name        string
		methodology string
		setupDur    int // for SetDuration
		expected    int
	}{
		{
			"strength work-only",
			"strength", 0,
			60, // single 60s activity
		},
		{
			"emom work-only",
			"emom", 0,
			300, // 5 rounds * 60s, no warmup/cooldown
		},
		{
			"for_time work-only",
			"for_time", 0,
			60, // single 60s activity, estimated like interval-based
		},
	}

	for _, tt := range tests {
		var routines []Routine
		switch tt.methodology {
		case "emom":
			routines = []Routine{routine("work", 0, []Block{
				block(5, 0, []Activity{act(0, 10, 0)}),
			})}
		default:
			routines = []Routine{routine("work", 0, []Block{
				block(1, 0, []Activity{act(60, 0, 0)}),
			})}
		}
		tr := makeTraining(tt.methodology, routines)
		got := tr.CalculateDuration()
		if got != tt.expected {
			t.Errorf("%s: CalculateDuration() = %d, want %d", tt.name, got, tt.expected)
		}
	}
}

func TestSetDuration_AMRAP_SkipWarmupCooldown(t *testing.T) {
	// amrap with no warmup/cooldown: full requested time goes to work
	tr := makeTraining("amrap", []Routine{
		routine("work", 0, []Block{
			block(1, 0, []Activity{act(0, 10, 0)}),
		}),
	})

	tr.SetDuration(20) // 20 min = 1200s, no warmup/cooldown to subtract

	if tr.Duration != 1200 {
		t.Errorf("amrap skip warmup/cooldown: Duration = %d, want 1200", tr.Duration)
	}
	if total := tr.CalculateDuration(); total != 1200 {
		t.Errorf("amrap skip warmup/cooldown: CalculateDuration() = %d, want 1200", total)
	}
}

func TestValidate_WeightModifierConsistency(t *testing.T) {
	validExercises := map[string]bool{"bench-press": true, "push-up": true}
	validModifiers := map[string]bool{"weight": true, "weighted vest": true}
	validRoutines := map[string]bool{"work": true}
	weightedModifiers := map[string]bool{"weight": true, "weighted vest": true}

	tests := []struct {
		name    string
		activity Activity
		wantErr bool
	}{
		{
			"weight_kg with weighted modifier passes",
			actWeighted("bench-press", 60, []string{"weight"}),
			false,
		},
		{
			"weight_kg with vest modifier passes",
			actWeighted("push-up", 10, []string{"weighted vest"}),
			false,
		},
		{
			"weight_kg without weighted modifier fails",
			actWeighted("bench-press", 60, nil),
			true,
		},
		{
			"vest modifier without weight_kg passes",
			actWeighted("push-up", 0, []string{"weighted vest"}),
			false,
		},
		{
			"bodyweight without modifier passes",
			actWeighted("push-up", 0, nil),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := Training{
				Name: "Test",
				Routines: []Routine{{
					Type: "work",
					Blocks: []Block{{
						Repeats:    1,
						Activities: []Activity{tt.activity},
					}},
				}},
			}
			err := tr.Validate(validExercises, validModifiers, validRoutines, weightedModifiers, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
