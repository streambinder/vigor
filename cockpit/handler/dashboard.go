package handler

import (
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/streambinder/vigor/cockpit/database"
	"github.com/streambinder/vigor/cockpit/view"
)

func Dashboard(c *fiber.Ctx) error {
	userCount, _ := database.GetUserCount()
	trainingCount, _ := database.GetTrainingCount()
	avgTrainingsPerDay, _ := database.GetAvgTrainingsPerDay()
	trainingStats, _ := database.GetTrainingGenerationStats(14)
	handlerStats, _ := database.GetHandlerRequestStats(14)
	errorStats, _ := database.GetHandlerErrorStats(14)
	trainings, _ := database.GetTrainings()
	reports, _ := database.GetReports()

	data := view.DashboardData{
		UserCount:                   userCount,
		TrainingCount:               trainingCount,
		AvgTrainingsPerDay:          avgTrainingsPerDay,
		TrainingGenerationLatencies: toLatencySeries(trainingStats),
		HandlerRequestLatencies:     toLatencySeries(handlerStats),
		HandlerRequestErrors:        toErrorSeries(errorStats),
		Trainings:                   trainings,
		Reports:                     reports,
	}

	return render(c, view.Dashboard(data))
}

func toLatencySeries(stats []database.LatencyPoint) []view.LatencySeries {
	// group points by Group field, preserving day order
	seriesMap := make(map[string][]view.LatencyDataPoint)
	var seriesOrder []string

	for _, s := range stats {
		label := s.Day
		if t, err := time.Parse("2006-01-02", s.Day); err == nil {
			label = t.Format("Jan 2")
		}
		if _, exists := seriesMap[s.Group]; !exists {
			seriesOrder = append(seriesOrder, s.Group)
		}
		seriesMap[s.Group] = append(seriesMap[s.Group], view.LatencyDataPoint{
			Label: label,
			Value: s.AvgMs,
		})
	}

	var series []view.LatencySeries
	for _, name := range seriesOrder {
		series = append(series, view.LatencySeries{
			Name:   name,
			Points: seriesMap[name],
		})
	}
	return series
}

func toErrorSeries(stats []database.ErrorPoint) []view.LatencySeries {
	seriesMap := make(map[string][]view.LatencyDataPoint)
	var seriesOrder []string

	for _, s := range stats {
		label := s.Day
		if t, err := time.Parse("2006-01-02", s.Day); err == nil {
			label = t.Format("Jan 2")
		}
		if _, exists := seriesMap[s.Group]; !exists {
			seriesOrder = append(seriesOrder, s.Group)
		}
		seriesMap[s.Group] = append(seriesMap[s.Group], view.LatencyDataPoint{
			Label: label,
			Value: float64(s.Count),
		})
	}

	var series []view.LatencySeries
	for _, name := range seriesOrder {
		series = append(series, view.LatencySeries{
			Name:   name,
			Points: seriesMap[name],
		})
	}
	return series
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
