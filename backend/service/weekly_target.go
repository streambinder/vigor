package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
)

const weeklyTargetHistoryWeeks = 4

// GetWeeklyTarget computes the weekly target for a user based on their goals and history.
func GetWeeklyTarget(userID uuid.UUID) (model.WeeklyTarget, error) {
	var profile model.Profile
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return model.WeeklyTarget{}, err
	}

	goalIDs := profile.Goals()
	if len(goalIDs) == 0 {
		return model.WeeklyTarget{Goals: []string{}}, nil
	}

	var goals []model.Goal
	if err := database.Knowledge.Where("id IN ?", goalIDs).Find(&goals).Error; err != nil {
		return model.WeeklyTarget{}, err
	}

	recommendation := SynthesizeRecommendations(goals)
	currentWeek := computeCurrentWeek(userID)
	history := computeHistory(userID, recommendation.SessionsPerWeek[0])

	return model.WeeklyTarget{
		Goals:          goalIDs,
		Recommendation: recommendation,
		CurrentWeek:    currentWeek,
		History:        history,
	}, nil
}

// SynthesizeRecommendations merges recommendations from multiple goals.
func SynthesizeRecommendations(goals []model.Goal) model.SynthesizedRecommendation {
	if len(goals) == 0 {
		return model.SynthesizedRecommendation{}
	}

	if len(goals) == 1 {
		g := goals[0]
		return model.SynthesizedRecommendation{
			SessionsPerWeek:     [2]int{int(g.SessionsPerWeek[0]), int(g.SessionsPerWeek[1])},
			SessionDurationMins: [2]int{int(g.SessionDurationMins[0]), int(g.SessionDurationMins[1])},
			MethodologyMix:      parseMethodologyWeights(g.MethodologyWeights),
			PreferredHours:      [2]int{int(g.PreferredHours[0]), int(g.PreferredHours[1])},
		}
	}

	var (
		minSessions, maxSessions int
		minDuration, maxDuration int
		minHour, maxHour         = 0, 24
		hasTimePreference        bool
		methodologySum           = make(map[string]float64)
	)

	for i, g := range goals {
		// sessions: take higher bounds
		if i == 0 || int(g.SessionsPerWeek[0]) > minSessions {
			minSessions = int(g.SessionsPerWeek[0])
		}
		if i == 0 || int(g.SessionsPerWeek[1]) > maxSessions {
			maxSessions = int(g.SessionsPerWeek[1])
		}

		// duration: find overlap
		if i == 0 {
			minDuration = int(g.SessionDurationMins[0])
			maxDuration = int(g.SessionDurationMins[1])
		} else {
			if int(g.SessionDurationMins[0]) > minDuration {
				minDuration = int(g.SessionDurationMins[0])
			}
			if int(g.SessionDurationMins[1]) < maxDuration {
				maxDuration = int(g.SessionDurationMins[1])
			}
		}

		// preferred hours: specific beats flexible, conflict → flexible
		isFlexible := g.PreferredHours[0] == 0 && g.PreferredHours[1] == 24
		if !isFlexible {
			if !hasTimePreference {
				minHour = int(g.PreferredHours[0])
				maxHour = int(g.PreferredHours[1])
				hasTimePreference = true
			} else if minHour != int(g.PreferredHours[0]) || maxHour != int(g.PreferredHours[1]) {
				minHour, maxHour = 0, 24
			}
		}

		// methodology: sum weights
		for method, weight := range g.MethodologyWeights {
			if w, ok := weight.(float64); ok {
				methodologySum[method] += w
			}
		}
	}

	// if duration overlap is invalid, average instead
	if minDuration > maxDuration {
		var sumMin, sumMax int
		for _, g := range goals {
			sumMin += int(g.SessionDurationMins[0])
			sumMax += int(g.SessionDurationMins[1])
		}
		minDuration = sumMin / len(goals)
		maxDuration = sumMax / len(goals)
	}

	// normalize methodology weights
	var total float64
	for _, w := range methodologySum {
		total += w
	}
	methodologyMix := make(map[string]float64)
	for method, w := range methodologySum {
		methodologyMix[method] = w / total
	}

	return model.SynthesizedRecommendation{
		SessionsPerWeek:     [2]int{minSessions, maxSessions},
		SessionDurationMins: [2]int{minDuration, maxDuration},
		MethodologyMix:      methodologyMix,
		PreferredHours:      [2]int{minHour, maxHour},
	}
}

func parseMethodologyWeights(weights map[string]any) map[string]float64 {
	result := make(map[string]float64)
	for k, v := range weights {
		if w, ok := v.(float64); ok {
			result[k] = w
		}
	}
	return result
}

func computeCurrentWeek(userID uuid.UUID) model.WeekProgress {
	now := time.Now()
	weekStart := startOfWeek(now)
	weekEnd := weekStart.AddDate(0, 0, 7)

	var trainings []model.Training
	database.DB.
		Where("(user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?)) AND completed_at >= ? AND completed_at < ? AND completed_at IS NOT NULL",
			userID, userID, weekStart, weekEnd).
		Find(&trainings)

	methodologies := make(map[string]bool)
	completedDays := make(map[int]bool)
	var totalDuration int

	for _, t := range trainings {
		if t.CompletedAt != nil {
			methodologies[t.Methodology] = true
			totalDuration += t.Duration
			dayOfWeek := int(t.CompletedAt.Weekday())
			// convert Sunday=0 to Sunday=6, Monday=1 to Monday=0
			if dayOfWeek == 0 {
				dayOfWeek = 6
			} else {
				dayOfWeek--
			}
			completedDays[dayOfWeek] = true
		}
	}

	var methodList []string
	for m := range methodologies {
		methodList = append(methodList, m)
	}

	var dayList []int
	for d := range completedDays {
		dayList = append(dayList, d)
	}

	avgDuration := 0
	if len(trainings) > 0 {
		avgDuration = (totalDuration / len(trainings)) / 60
	}

	daysRemaining := max(int(weekEnd.Sub(now).Hours()/24), 0)

	return model.WeekProgress{
		WeekStart:         weekStart,
		SessionsCompleted: len(trainings),
		DaysRemaining:     daysRemaining,
		MethodologiesUsed: methodList,
		AvgDurationMins:   avgDuration,
		CompletedDays:     dayList,
	}
}

func computeHistory(userID uuid.UUID, minTarget int) []model.WeekSummary {
	now := time.Now()
	currentWeekStart := startOfWeek(now)

	var history []model.WeekSummary

	for i := 1; i <= weeklyTargetHistoryWeeks; i++ {
		weekStart := currentWeekStart.AddDate(0, 0, -7*i)
		weekEnd := weekStart.AddDate(0, 0, 7)

		var count int64
		database.DB.Model(&model.Training{}).
			Where("(user_id = ? OR id IN (SELECT training_id FROM partners WHERE user_id = ?)) AND completed_at >= ? AND completed_at < ? AND completed_at IS NOT NULL",
				userID, userID, weekStart, weekEnd).
			Count(&count)

		history = append(history, model.WeekSummary{
			WeekStart:         weekStart,
			SessionsCompleted: int(count),
			OnTarget:          int(count) >= minTarget,
		})
	}

	return history
}

// startOfWeek returns Monday 00:00:00 of the week containing t.
func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-weekday+1, 0, 0, 0, 0, t.Location())
}
