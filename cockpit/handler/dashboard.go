package handler

import (
	"sort"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/streambinder/vigor/cockpit/database"
	"github.com/streambinder/vigor/cockpit/view"
)

func Dashboard(c *fiber.Ctx) error {
	userCount, _ := database.GetUserCount()
	avgActiveUsersPerDay, _ := database.GetAvgActiveUsersPerDay()
	trainingGenerationCount, _ := database.GetTrainingGenerationCount()
	trainingDbCount, _ := database.GetTrainingCount()
	if trainingDbCount > trainingGenerationCount {
		trainingGenerationCount = trainingDbCount
	}
	avgTrainingGenerationsPerDay, _ := database.GetAvgTrainingGenerationsPerDay()
	completedTrainingCount, _ := database.GetCompletedTrainingCount()
	avgCompletedTrainingsPerDay, _ := database.GetAvgCompletedTrainingsPerDay()
	activeUsersStats, _ := database.GetActiveUsersPerDay(14)
	activeUsersPerUserStats, _ := database.GetActiveUsersPerDayPerUser(7)
	topActiveUsersStats, _ := database.GetTopActiveUsers(10)
	trainingStats, _ := database.GetTrainingGenerationStats(14)
	handlerStats, _ := database.GetHandlerRequestStats(14)
	errorStats, _ := database.GetHandlerErrorStats(14)
	trainingFailureStats, _ := database.GetTrainingGenerationFailures(14)
	badQualityStats, _ := database.GetBadQualityPerModel()
	trainings, _ := database.GetTrainings()
	reports, _ := database.GetReports()
	users, _ := database.GetUsers()

	data := view.DashboardData{
		UserCount:                    userCount,
		AvgActiveUsersPerDay:         avgActiveUsersPerDay,
		TrainingGenerationCount:      trainingGenerationCount,
		AvgTrainingGenerationsPerDay: avgTrainingGenerationsPerDay,
		CompletedTrainingCount:       completedTrainingCount,
		AvgCompletedTrainingsPerDay:  avgCompletedTrainingsPerDay,
		ActiveUsersPerDay:            toActiveUsersSeries(activeUsersStats),
		ActiveUsersPerDayPerUser:     toActiveUsersPerUserSeries(activeUsersPerUserStats),
		TopActiveUsers:               toTopActiveUsers(topActiveUsersStats),
		TrainingGenerationLatencies:  toLatencySeries(trainingStats),
		HandlerRequestLatencies:      toLatencySeries(handlerStats),
		HandlerRequestErrors:         toErrorSeries(errorStats),
		TrainingGenerationFailures:   toErrorSeries(trainingFailureStats),
		BadQualityPerModel:           toBadQualitySeries(badQualityStats),
		Trainings:                    trainings,
		Reports:                      reports,
		Users:                        users,
	}

	return render(c, view.Dashboard(data))
}

func toLatencySeries(stats []database.LatencyPoint) []view.LatencySeries {
	if len(stats) == 0 {
		return nil
	}

	// collect unique days sorted
	daySet := make(map[string]bool)
	for _, s := range stats {
		daySet[s.Day] = true
	}
	var allDays []string
	for day := range daySet {
		allDays = append(allDays, day)
	}
	sort.Strings(allDays)

	// group points by Group field
	seriesMap := make(map[string]map[string]float64)
	var seriesOrder []string
	for _, s := range stats {
		if _, exists := seriesMap[s.Group]; !exists {
			seriesMap[s.Group] = make(map[string]float64)
			seriesOrder = append(seriesOrder, s.Group)
		}
		seriesMap[s.Group][s.Day] = s.P95Ms
	}

	// build series with only actual data points (no zero-filling)
	var series []view.LatencySeries
	for _, name := range seriesOrder {
		var points []view.LatencyDataPoint
		dayValues := seriesMap[name]
		for _, day := range allDays {
			if val, exists := dayValues[day]; exists {
				label := day
				if t, err := time.Parse("2006-01-02", day); err == nil {
					label = t.Format("Jan 2")
				}
				points = append(points, view.LatencyDataPoint{
					Label: label,
					Value: val,
				})
			}
		}
		series = append(series, view.LatencySeries{Name: name, Points: points})
	}
	return series
}

func toErrorSeries(stats []database.ErrorPoint) []view.LatencySeries {
	if len(stats) == 0 {
		return nil
	}

	// collect all days and find date range
	daySet := make(map[string]bool)
	for _, s := range stats {
		daySet[s.Day] = true
	}
	allDays := generateDayRange(daySet)

	// group points by Group field
	seriesMap := make(map[string]map[string]float64)
	var seriesOrder []string
	for _, s := range stats {
		if _, exists := seriesMap[s.Group]; !exists {
			seriesMap[s.Group] = make(map[string]float64)
			seriesOrder = append(seriesOrder, s.Group)
		}
		seriesMap[s.Group][s.Day] = float64(s.Count)
	}

	// build series with zero-filled gaps
	var series []view.LatencySeries
	for _, name := range seriesOrder {
		var points []view.LatencyDataPoint
		dayValues := seriesMap[name]
		for _, day := range allDays {
			label := day
			if t, err := time.Parse("2006-01-02", day); err == nil {
				label = t.Format("Jan 2")
			}
			points = append(points, view.LatencyDataPoint{
				Label: label,
				Value: dayValues[day],
			})
		}
		series = append(series, view.LatencySeries{Name: name, Points: points})
	}
	return series
}

func toActiveUsersSeries(stats []database.ActiveUsersPoint) []view.LatencySeries {
	if len(stats) == 0 {
		return nil
	}

	// collect all days and find date range
	daySet := make(map[string]bool)
	dayValues := make(map[string]float64)
	for _, s := range stats {
		daySet[s.Day] = true
		dayValues[s.Day] = float64(s.Count)
	}
	allDays := generateDayRange(daySet)

	// build series with zero-filled gaps
	var points []view.LatencyDataPoint
	for _, day := range allDays {
		label := day
		if t, err := time.Parse("2006-01-02", day); err == nil {
			label = t.Format("Jan 2")
		}
		points = append(points, view.LatencyDataPoint{
			Label: label,
			Value: dayValues[day],
		})
	}
	return []view.LatencySeries{{Name: "Active Users", Points: points}}
}

func toActiveUsersPerUserSeries(stats []database.ActiveUserPoint) []view.LatencySeries {
	if len(stats) == 0 {
		return nil
	}

	// collect unique user IDs for name resolution
	userIDSet := make(map[string]bool)
	for _, s := range stats {
		userIDSet[s.UserID] = true
	}
	var userIDs []string
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	// resolve names from postgres
	userNames, _ := database.GetUserNames(userIDs)

	// collect all days and find date range
	daySet := make(map[string]bool)
	for _, s := range stats {
		daySet[s.Day] = true
	}
	allDays := generateDayRange(daySet)

	// group points by user
	seriesMap := make(map[string]map[string]float64)
	var seriesOrder []string
	for _, s := range stats {
		if _, exists := seriesMap[s.UserID]; !exists {
			seriesMap[s.UserID] = make(map[string]float64)
			seriesOrder = append(seriesOrder, s.UserID)
		}
		seriesMap[s.UserID][s.Day] = float64(s.Count)
	}

	// build series with zero-filled gaps
	var series []view.LatencySeries
	for _, userID := range seriesOrder {
		name := userNames[userID]
		if name == "" {
			name = userID[:8]
		}
		var points []view.LatencyDataPoint
		dayValues := seriesMap[userID]
		for _, day := range allDays {
			label := day
			if t, err := time.Parse("2006-01-02", day); err == nil {
				label = t.Format("Jan 2")
			}
			points = append(points, view.LatencyDataPoint{
				Label: label,
				Value: dayValues[day],
			})
		}
		series = append(series, view.LatencySeries{Name: name, Points: points})
	}
	return series
}

func toTopActiveUsers(stats []database.UserActivityRank) []view.TopActiveUser {
	if len(stats) == 0 {
		return nil
	}

	// resolve names
	var userIDs []string
	for _, s := range stats {
		userIDs = append(userIDs, s.UserID)
	}
	userNames, _ := database.GetUserNames(userIDs)

	result := make([]view.TopActiveUser, len(stats))
	for i, s := range stats {
		name := userNames[s.UserID]
		if name == "" {
			name = s.UserID[:8]
		}
		result[i] = view.TopActiveUser{
			Name:       name,
			ActiveDays: s.ActiveDays,
			TotalDays:  s.TotalDays,
			Percentage: s.Percentage,
		}
	}
	return result
}

// generateDayRange returns all days between min date found in daySet and today
func generateDayRange(daySet map[string]bool) []string {
	if len(daySet) == 0 {
		return nil
	}

	now := time.Now()
	today, _ := time.Parse("2006-01-02", now.Format("2006-01-02"))
	var minDay time.Time
	first := true
	for day := range daySet {
		t, err := time.Parse("2006-01-02", day)
		if err != nil {
			continue
		}
		if first || t.Before(minDay) {
			minDay = t
			first = false
		}
	}

	var days []string
	for d := minDay; !d.After(today); d = d.AddDate(0, 0, 1) {
		days = append(days, d.Format("2006-01-02"))
	}
	return days
}

func toBadQualitySeries(stats []database.ModelQualityPoint) []view.LatencySeries {
	if len(stats) == 0 {
		return nil
	}
	points := make([]view.LatencyDataPoint, len(stats))
	for i, s := range stats {
		points[i] = view.LatencyDataPoint{Label: s.Model, Value: float64(s.Count)}
	}
	return []view.LatencySeries{{Name: "Bad Quality", Points: points}}
}

func DeleteReport(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := database.DeleteReport(id); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return adaptor.HTTPHandler(templ.Handler(component))(c)
}
