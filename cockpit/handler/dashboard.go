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
	trainingStats, _ := database.GetTrainingGenerationStats(14)
	handlerStats, _ := database.GetHandlerRequestStats(14)
	errorStats, _ := database.GetHandlerErrorStats(14)
	trainingFailureStats, _ := database.GetTrainingGenerationFailures(14)
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
		TrainingGenerationLatencies:  toLatencySeries(trainingStats),
		HandlerRequestLatencies:      toLatencySeries(handlerStats),
		HandlerRequestErrors:         toErrorSeries(errorStats),
		TrainingGenerationFailures:   toErrorSeries(trainingFailureStats),
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
