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

func actWeighted(exerciseID string, weightKg float64, modifiers []string) Activity {
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

func TestCalculateDuration_EMOM_WithBlockRest(t *testing.T) {
	tr := makeTraining("emom", []Routine{
		routine("work", 0, []Block{
			block(6, 120, []Activity{act(0, 10, 0)}), // 6×60 = 360s + 120s rest
			block(6, 120, []Activity{act(0, 10, 0)}), // 6×60 = 360s + 120s rest
			block(5, 120, []Activity{act(0, 10, 0)}), // 5×60 = 300s (last block, no rest)
		}),
	})

	got := tr.CalculateDuration()

	// work: (6+6+5)×60 = 1020s
	// inter-block rest: 120 + 120 = 240s (last block excluded)
	// total = 1260s

	if got != 1260 {
		t.Errorf("emom with block rest CalculateDuration() = %d, want 1260", got)
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

func TestSetDuration_Strength_ScalesRepeats(t *testing.T) {
	// single work block, 3 repeats, ~450s of work (see TestCalculateDuration_Strength)
	// request 15 min (900s) → should roughly double repeats
	tr := makeTraining("strength", []Routine{
		routine("warmup", 0, []Block{
			block(1, 0, []Activity{act(60, 0, 0)}),
		}),
		routine("work", 0, []Block{
			block(3, 60, []Activity{
				act(0, 10, 30), // 40s work
				act(0, 10, 0),  // 40s work
			}),
		}),
		routine("cooldown", 0, []Block{
			block(1, 0, []Activity{act(60, 0, 0)}),
		}),
	})

	tr.SetDuration(15) // 15 min = 900s total, warmup=60 cooldown=60 → 780s work budget
	// iterative scaling: starts at 3 repeats, adds +1 until closest to 900s total
	if tr.Routines[1].Blocks[0].Repeats != 5 {
		t.Errorf("strength SetDuration(15): repeats = %d, want 5", tr.Routines[1].Blocks[0].Repeats)
	}

	// verify duration is within tolerance of target
	total := tr.CalculateDuration()
	if total < 765 || total > 1035 { // 900 ± 15%
		t.Errorf("strength SetDuration(15): total = %d, want within 765-1035", total)
	}
}

func TestSetDuration_Strength_ScalesDown(t *testing.T) {
	// work block with 6 repeats, request a short duration to verify scaling down
	tr := makeTraining("strength", []Routine{
		routine("work", 0, []Block{
			block(6, 60, []Activity{
				act(0, 10, 30),
				act(0, 10, 0),
			}),
		}),
	})

	tr.SetDuration(5) // 5 min = 300s, no warmup/cooldown → 300s work budget
	// iterative scaling: starts at 6 repeats (960s), removes -1 until closest to 300s
	if tr.Routines[0].Blocks[0].Repeats != 2 {
		t.Errorf("strength SetDuration(5): repeats = %d, want 2", tr.Routines[0].Blocks[0].Repeats)
	}
}

func TestSetDuration_Strength_ClampsToOne(t *testing.T) {
	// request very short duration → repeats should clamp to 1, never 0
	tr := makeTraining("strength", []Routine{
		routine("work", 0, []Block{
			block(10, 60, []Activity{
				act(0, 10, 30),
				act(0, 10, 0),
			}),
		}),
	})

	tr.SetDuration(1) // 1 min = 60s → multiplier is very small
	if tr.Routines[0].Blocks[0].Repeats < 1 {
		t.Errorf("strength SetDuration(1): repeats = %d, want >= 1", tr.Routines[0].Blocks[0].Repeats)
	}
}

func TestSetDuration_EMOM_ScalesRepeats(t *testing.T) {
	tr := makeTraining("emom", []Routine{
		routine("warmup", 0, []Block{
			block(1, 0, []Activity{act(120, 0, 0)}),
		}),
		routine("work", 0, []Block{
			block(10, 0, []Activity{act(0, 5, 0)}), // 10 rounds
			block(5, 0, []Activity{act(0, 8, 0)}),  // 5 rounds
		}),
		routine("cooldown", 0, []Block{
			block(1, 0, []Activity{act(60, 0, 0)}),
		}),
	})

	// original: warmup=120 + work=15*60=900 + cooldown=60 = 1080s = 18m
	// request 30m (1800s) → work budget = 1620s → 27 total repeats needed
	tr.SetDuration(30)

	totalRepeats := tr.Routines[1].Blocks[0].Repeats + tr.Routines[1].Blocks[1].Repeats
	if totalRepeats != 27 {
		t.Errorf("emom SetDuration(30): total repeats = %d, want 27", totalRepeats)
	}

	total := tr.CalculateDuration()
	// warmup(120) + 27*60(1620) + cooldown(60) = 1800
	if total != 1800 {
		t.Errorf("emom SetDuration(30): total = %d, want 1800", total)
	}
}

func TestSetDuration_EMOM_ScalesDown(t *testing.T) {
	tr := makeTraining("emom", []Routine{
		routine("work", 0, []Block{
			block(20, 0, []Activity{act(0, 5, 0)}),
		}),
	})

	// original: 20*60=1200s=20m. request 10m → 10 target repeats
	tr.SetDuration(10)

	if tr.Routines[0].Blocks[0].Repeats != 10 {
		t.Errorf("emom SetDuration(10): repeats = %d, want 10", tr.Routines[0].Blocks[0].Repeats)
	}
}

func TestSetDuration_EMOM_RoundingDriftCorrection(t *testing.T) {
	// 3 blocks with repeats that cause rounding drift
	tr := makeTraining("emom", []Routine{
		routine("work", 0, []Block{
			block(3, 0, []Activity{act(0, 5, 0)}), // 3 rounds
			block(3, 0, []Activity{act(0, 5, 0)}), // 3 rounds
			block(3, 0, []Activity{act(0, 5, 0)}), // 3 rounds → total 9
		}),
	})

	// request 10m → 10 target repeats, ratio = 10/9 ≈ 1.11
	// round(3*1.11) = round(3.33) = 3 per block → 9 assigned, need 10
	// drift correction adds 1 to last block → 3, 3, 4
	tr.SetDuration(10)

	totalRepeats := 0
	for _, b := range tr.Routines[0].Blocks {
		totalRepeats += b.Repeats
	}
	if totalRepeats != 10 {
		t.Errorf("emom rounding drift: total repeats = %d, want 10", totalRepeats)
	}
}

func TestSetDuration_MultipleWorkBlocks(t *testing.T) {
	// two work blocks with different repeats, verify duration accuracy
	tr := makeTraining("circuit", []Routine{
		routine("work", 0, []Block{
			block(2, 30, []Activity{act(30, 0, 10), act(30, 0, 0)}), // 2 repeats
			block(4, 30, []Activity{act(30, 0, 10), act(30, 0, 0)}), // 4 repeats
		}),
	})

	// compute original work duration, then request double
	originalWork := tr.calcIntervalDuration("work")
	requestedMinutes := (originalWork * 2) / 60

	tr.SetDuration(requestedMinutes)

	// verify duration is within 15% of target
	target := requestedMinutes * 60
	total := tr.CalculateDuration()
	low := float64(target) * 0.85
	high := float64(target) * 1.15
	if float64(total) < low || float64(total) > high {
		t.Errorf("circuit SetDuration: total = %d, want within %.0f-%.0f", total, low, high)
	}

	// both blocks should still have at least 1 repeat
	if tr.Routines[0].Blocks[0].Repeats < 1 || tr.Routines[0].Blocks[1].Repeats < 1 {
		t.Errorf("circuit SetDuration: block repeats below 1: %d, %d",
			tr.Routines[0].Blocks[0].Repeats, tr.Routines[0].Blocks[1].Repeats)
	}
}

func TestSetDuration_DifferentPerRepeatCosts(t *testing.T) {
	// block 1: cheap per repeat (short activity, no rest) → 30s/repeat
	// block 2: expensive per repeat (long activity + rest) → 120s+60s=180s/repeat
	// the iterative approach should still land within tolerance
	tr := makeTraining("strength", []Routine{
		routine("warmup", 0, []Block{
			block(1, 0, []Activity{act(60, 0, 0)}),
		}),
		routine("work", 0, []Block{
			block(2, 0, []Activity{act(30, 0, 0)}),  // 2 × 30s = 60s
			block(2, 60, []Activity{act(120, 0, 0)}), // 2 × (120+60) - 60 last rest suppressed = 300s
		}),
		routine("cooldown", 0, []Block{
			block(1, 0, []Activity{act(60, 0, 0)}),
		}),
	})

	tr.SetDuration(15) // 15min = 900s

	total := tr.CalculateDuration()
	target := 900
	low := float64(target) * (1 - durationTolerancePct)
	high := float64(target) * (1 + durationTolerancePct)
	if float64(total) < low || float64(total) > high {
		t.Errorf("different costs SetDuration(15): total = %d, want within %.0f-%.0f", total, low, high)
	}
}

func TestSetDuration_StopsWhenNoImprovement(t *testing.T) {
	// single block: 1 repeat=600s, 2 repeats=1200s. target=700s.
	// +1 gives |1200-700|=500 > |600-700|=100, so algorithm should stay at 1.
	tr := makeTraining("strength", []Routine{
		routine("work", 0, []Block{
			block(1, 0, []Activity{act(600, 0, 0)}),
		}),
	})

	tr.SetDuration(700 / 60) // ~11min

	if tr.Routines[0].Blocks[0].Repeats != 1 {
		t.Errorf("expected 1 repeat (closer to target), got %d", tr.Routines[0].Blocks[0].Repeats)
	}
}

func TestSetDuration_BalancesBlockRepeats(t *testing.T) {
	// two superset blocks with different per-repeat costs:
	// block 0: cheap (~100s/repeat), block 1: expensive (~180s/repeat)
	// without balancing, the cheap block would get all the increments (e.g. 3→12)
	// while the expensive one stays at 3. the algorithm should distribute evenly.
	tr := makeTraining("supersets", []Routine{
		routine("work", 0, []Block{
			block(3, 60, []Activity{
				act(0, 8, 0),  // ~32s work
				act(0, 10, 60), // ~40s work + 60s rest
			}),
			block(3, 60, []Activity{
				act(0, 10, 0),  // ~40s work
				act(0, 12, 60), // ~48s work + 60s rest
			}),
		}),
	})

	tr.SetDuration(35) // 35 min = 2100s

	r0 := tr.Routines[0].Blocks[0].Repeats
	r1 := tr.Routines[0].Blocks[1].Repeats

	// both started at 3, so proportional scaling = equal scaling.
	// blocks should be within 2 repeats of each other.
	diff := r0 - r1
	if diff < 0 {
		diff = -diff
	}
	if diff > 2 {
		t.Errorf("blocks unbalanced: block0=%d, block1=%d (diff=%d, want <= 2)", r0, r1, diff)
	}

	// verify duration is still within tolerance
	total := tr.CalculateDuration()
	target := 2100
	low := float64(target) * (1 - durationTolerancePct)
	high := float64(target) * (1 + durationTolerancePct)
	if float64(total) < low || float64(total) > high {
		t.Errorf("balanced SetDuration(35): total = %d, want within %.0f-%.0f", total, low, high)
	}
}

func TestSetDuration_PreservesRepeatRatio(t *testing.T) {
	// block 0 starts at 2 repeats, block 1 starts at 6 repeats (1:3 ratio).
	// after scaling up, the ratio should be roughly preserved — block 1 should
	// always have more repeats than block 0.
	tr := makeTraining("supersets", []Routine{
		routine("work", 0, []Block{
			block(2, 60, []Activity{
				act(0, 8, 0),
				act(0, 10, 60),
			}),
			block(6, 60, []Activity{
				act(0, 10, 0),
				act(0, 12, 60),
			}),
		}),
	})

	tr.SetDuration(45) // 45 min = 2700s

	r0 := tr.Routines[0].Blocks[0].Repeats
	r1 := tr.Routines[0].Blocks[1].Repeats

	if r1 <= r0 {
		t.Errorf("ratio not preserved: block0=%d, block1=%d (block1 should be > block0 since it started 6 vs 2)", r0, r1)
	}
}

func TestSetDuration_WithinTolerance(t *testing.T) {
	// regression test: various targets should produce durations within 15% tolerance
	tests := []struct {
		name        string
		methodology string
		routines    []Routine
		minutes     int
	}{
		{
			"strength 45min",
			"strength",
			[]Routine{
				routine("warmup", 0, []Block{block(1, 0, []Activity{act(120, 0, 0)})}),
				routine("work", 0, []Block{
					block(3, 90, []Activity{act(0, 10, 30), act(0, 8, 0)}),
					block(3, 90, []Activity{act(0, 12, 30), act(0, 10, 0)}),
					block(3, 60, []Activity{act(0, 10, 0)}),
				}),
				routine("cooldown", 0, []Block{block(1, 0, []Activity{act(120, 0, 0)})}),
			},
			45,
		},
		{
			"circuit 30min",
			"circuit",
			[]Routine{
				routine("warmup", 0, []Block{block(1, 0, []Activity{act(60, 0, 0)})}),
				routine("work", 0, []Block{
					block(4, 30, []Activity{act(30, 0, 10), act(30, 0, 10), act(30, 0, 0)}),
				}),
				routine("cooldown", 0, []Block{block(1, 0, []Activity{act(60, 0, 0)})}),
			},
			30,
		},
		{
			"emom 20min",
			"emom",
			[]Routine{
				routine("warmup", 0, []Block{block(1, 0, []Activity{act(120, 0, 0)})}),
				routine("work", 0, []Block{
					block(5, 0, []Activity{act(0, 5, 0)}),
					block(5, 0, []Activity{act(0, 8, 0)}),
				}),
				routine("cooldown", 0, []Block{block(1, 0, []Activity{act(60, 0, 0)})}),
			},
			20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := makeTraining(tt.methodology, tt.routines)
			tr.SetDuration(tt.minutes)

			total := tr.CalculateDuration()
			target := tt.minutes * 60
			low := float64(target) * (1 - durationTolerancePct)
			high := float64(target) * (1 + durationTolerancePct)
			if float64(total) < low || float64(total) > high {
				t.Errorf("%s: total = %ds (%.1fm), want within %.0f-%.0f",
					tt.name, total, float64(total)/60, low, high)
			}
		})
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
	validExercises := map[string]string{"bench-press": "bench-press", "push-up": "push-up"}
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
				Name:        "Test",
				Methodology: "strength",
				Routines: []Routine{{
					Type: "work",
					Blocks: []Block{{
						Repeats:    1,
						Activities: []Activity{tt.activity},
					}},
				}},
			}
			err := tr.Validate(validExercises, validModifiers, validRoutines, weightedModifiers, nil, false, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate_DurationTolerance(t *testing.T) {
	validExercises := map[string]string{"ex1": "ex1"}
	validModifiers := map[string]bool{}
	validRoutines := map[string]bool{"work": true}
	weightedModifiers := map[string]bool{}

	makeTestTraining := func() Training {
		return Training{
			Name:        "Test",
			Methodology: "strength",
			Routines: []Routine{routine("work", 0, []Block{
				block(2, 60, []Activity{
					{ExerciseID: "ex1", Reps: 10, Rest: 30},
					{ExerciseID: "ex1", Reps: 10, Rest: 0},
				}),
				block(2, 60, []Activity{
					{ExerciseID: "ex1", Reps: 10, Rest: 30},
					{ExerciseID: "ex1", Reps: 10, Rest: 0},
				}),
				block(2, 60, []Activity{
					{ExerciseID: "ex1", Reps: 10, Rest: 30},
					{ExerciseID: "ex1", Reps: 10, Rest: 0},
				}),
			})},
		}
	}

	tr := makeTestTraining()
	actualDuration := tr.CalculateDuration()
	actualMinutes := actualDuration / 60 // 14m

	tests := []struct {
		name             string
		requestedMinutes int
		wantErr          bool
	}{
		{"exact match", actualMinutes, false},
		{"within tolerance (slightly over)", actualMinutes + 2, false},
		{"within tolerance (slightly under)", actualMinutes - 2, false},
		// SetDuration now scales repeats, so large upward mismatches get corrected
		{"way over requested (scaled)", actualMinutes * 3, false},
		// scaling down hits the 1-repeat clamp — rest times alone exceed 5m budget
		{"way under requested (clamped)", actualMinutes / 3, true},
		{"zero skips check", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := makeTestTraining()
			if err := tr.Validate(validExercises, validModifiers, validRoutines, weightedModifiers, nil, false, nil, nil); err != nil {
				t.Fatalf("Validate() unexpected structural error: %v", err)
			}
			tr.SetDuration(tt.requestedMinutes)
			err := tr.ValidateDuration(tt.requestedMinutes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDuration() with requested=%dm, actual=%ds: error = %v, wantErr %v",
					tt.requestedMinutes, actualDuration, err, tt.wantErr)
			}
		})
	}
}

func TestValidate_TargetMuscles(t *testing.T) {
	validExercises := map[string]string{"ex1": "ex1", "ex2": "ex2"}
	validModifiers := map[string]bool{}
	validRoutines := map[string]bool{"work": true}
	weightedModifiers := map[string]bool{}

	base := func() Training {
		return Training{
			Name:        "Test",
			Methodology: "strength",
			Routines: []Routine{routine("work", 0, []Block{
				block(1, 0, []Activity{
					{ExerciseID: "ex1", Reps: 10},
					{ExerciseID: "ex2", Reps: 10},
				}),
			})},
		}
	}

	tests := []struct {
		name    string
		target  []string
		actual  []string
		wantErr bool
	}{
		{"all targets covered", []string{"chest", "arms"}, []string{"chest", "arms", "core"}, false},
		{"exact match", []string{"chest", "arms"}, []string{"chest", "arms"}, false},
		{"missing target muscle", []string{"chest", "arms", "legs"}, []string{"chest", "arms"}, true},
		{"nil targets skips check", nil, []string{"chest"}, false},
		{"empty targets skips check", []string{}, []string{"chest"}, false},
		{"nil actuals skips check", []string{"chest"}, nil, false},
		{"empty actuals skips check", []string{"chest"}, []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := base()
			err := tr.Validate(validExercises, validModifiers, validRoutines, weightedModifiers, nil, false, tt.target, tt.actual)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate_NewChecks(t *testing.T) {
	validExercises := map[string]string{"ex1": "ex1"}
	validModifiers := map[string]bool{}
	validRoutines := map[string]bool{"work": true}
	weightedModifiers := map[string]bool{}

	// base valid training — all new checks pass
	base := func() Training {
		return Training{
			Name:        "Test",
			Methodology: "strength",
			Routines: []Routine{routine("work", 0, []Block{
				block(1, 0, []Activity{{ExerciseID: "ex1", Reps: 10}}),
			})},
		}
	}

	tests := []struct {
		name     string
		mutate   func(*Training)
		wantCode string
	}{
		{
			"valid training passes",
			func(tr *Training) {},
			"",
		},
		{
			"zero repeats",
			func(tr *Training) { tr.Routines[0].Blocks[0].Repeats = 0 },
			"zero_repeats",
		},
		{
			"negative repeats",
			func(tr *Training) { tr.Routines[0].Blocks[0].Repeats = -1 },
			"zero_repeats",
		},
		{
			"invalid methodology",
			func(tr *Training) { tr.Methodology = "crossfit" },
			"invalid_methodology",
		},
		{
			"empty methodology",
			func(tr *Training) { tr.Methodology = "" },
			"invalid_methodology",
		},
		{
			"no duration or reps",
			func(tr *Training) {
				tr.Routines[0].Blocks[0].Activities[0].Reps = 0
				tr.Routines[0].Blocks[0].Activities[0].Duration = 0
			},
			"no_duration_or_reps",
		},
		{
			"negative duration",
			func(tr *Training) { tr.Routines[0].Blocks[0].Activities[0].Duration = -1 },
			"negative_duration",
		},
		{
			"negative reps",
			func(tr *Training) {
				tr.Routines[0].Blocks[0].Activities[0].Reps = -1
				tr.Routines[0].Blocks[0].Activities[0].Duration = 10 // avoid no_duration_or_reps
			},
			"negative_reps",
		},
		{
			"negative activity rest",
			func(tr *Training) { tr.Routines[0].Blocks[0].Activities[0].Rest = -1 },
			"negative_activity_rest",
		},
		{
			"negative block rest",
			func(tr *Training) { tr.Routines[0].Blocks[0].Rest = -1 },
			"negative_block_rest",
		},
		{
			"negative routine rest",
			func(tr *Training) { tr.Routines[0].Rest = -1 },
			"negative_routine_rest",
		},
		{
			"negative weight",
			func(tr *Training) { tr.Routines[0].Blocks[0].Activities[0].WeightKg = -1 },
			"negative_weight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := base()
			tt.mutate(&tr)
			err := tr.Validate(validExercises, validModifiers, validRoutines, weightedModifiers, nil, false, nil, nil)
			if tt.wantCode == "" {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() expected error with code %q, got nil", tt.wantCode)
			}
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("Validate() error is not *ValidationError: %T", err)
			}
			if ve.Code != tt.wantCode {
				t.Errorf("Validate() error code = %q, want %q", ve.Code, tt.wantCode)
			}
		})
	}
}

func TestPurgeRepsDuration(t *testing.T) {
	tests := []struct {
		name         string
		methodology  string
		duration     int
		reps         int
		wantDuration int
		wantReps     int
	}{
		{"hiit: both set, keep duration", "hiit", 30, 15, 30, 0},
		{"circuit: both set, keep duration", "circuit", 40, 10, 40, 0},
		{"amrap: both set, keep duration", "amrap", 60, 20, 60, 0},
		{"for_time: both set, keep duration", "for_time", 45, 12, 45, 0},
		{"endurance: both set, keep duration", "endurance", 90, 8, 90, 0},
		{"mobility: both set, keep duration", "mobility", 30, 10, 30, 0},
		{"strength: both set, keep reps", "strength", 30, 10, 0, 10},
		{"supersets: both set, keep reps", "supersets", 40, 12, 0, 12},
		{"emom: both set, keep reps", "emom", 30, 8, 0, 8},
		{"hiit: only duration, no change", "hiit", 30, 0, 30, 0},
		{"strength: only reps, no change", "strength", 0, 10, 0, 10},
		{"hiit: only reps, no change", "hiit", 0, 15, 0, 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := makeTraining(tt.methodology, []Routine{
				routine("work", 0, []Block{
					block(1, 0, []Activity{{Duration: tt.duration, Reps: tt.reps}}),
				}),
			})
			tr.PurgeRepsDuration()
			a := tr.Routines[0].Blocks[0].Activities[0]
			if a.Duration != tt.wantDuration {
				t.Errorf("duration = %d, want %d", a.Duration, tt.wantDuration)
			}
			if a.Reps != tt.wantReps {
				t.Errorf("reps = %d, want %d", a.Reps, tt.wantReps)
			}
		})
	}
}
