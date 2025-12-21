package handler

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/streambinder/vigor/cockpit/database"
	"github.com/streambinder/vigor/cockpit/view"
)

func Dashboard(c *fiber.Ctx) error {
	userCount, _ := database.GetUserCount()
	trainingCount, _ := database.GetTrainingCount()
	latencyStats, _ := database.GetLatencyStats(14)

	var latencies []view.LatencyDataPoint
	for _, s := range latencyStats {
		latencies = append(latencies, view.LatencyDataPoint{
			Label: s.Day.Format("Jan 2"),
			Value: s.AvgMs,
		})
	}

	data := view.DashboardData{
		UserCount:     userCount,
		TrainingCount: trainingCount,
		Latencies:     latencies,
	}

	return render(c, view.Dashboard(data))
}

func render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return adaptor.HTTPHandler(templ.Handler(component))(c)
}
