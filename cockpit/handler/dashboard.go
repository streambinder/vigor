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
	latencyStats, _ := database.GetLatencyStats(14)
	dbReports, _ := database.GetReports()

	var latencies []view.LatencyDataPoint
	for _, s := range latencyStats {
		label := s.Day
		if t, err := time.Parse("2006-01-02", s.Day); err == nil {
			label = t.Format("Jan 2")
		}
		latencies = append(latencies, view.LatencyDataPoint{
			Label: label,
			Value: s.AvgMs,
		})
	}

	var reports []view.Report
	for _, r := range dbReports {
		reports = append(reports, view.Report{
			ID:         r.ID,
			Content:    r.Content,
			TrainingID: r.TrainingID,
			ActivityID: r.ActivityID,
			UserID:     r.UserID,
			UserEmail:  r.UserEmail,
			CreatedAt:  r.CreatedAt,
		})
	}

	data := view.DashboardData{
		UserCount:          userCount,
		TrainingCount:      trainingCount,
		AvgTrainingsPerDay: avgTrainingsPerDay,
		Latencies:          latencies,
		Reports:            reports,
	}

	return render(c, view.Dashboard(data))
}

func render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return adaptor.HTTPHandler(templ.Handler(component))(c)
}
