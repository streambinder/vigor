package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/google/uuid"
	"github.com/streambinder/vigor/handler/middleware"
	"github.com/streambinder/vigor/model"
	"github.com/streambinder/vigor/service"
)

func initHealth(app *fiber.App) {
	// budget covers one full incremental sync (worst case ~30 per-date batches);
	// second logical sync inside the window is blocked → effective 1 sync/min/user
	syncLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			if uid, ok := c.Locals("userID").(uuid.UUID); ok {
				return "health_sync:" + uid.String()
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limit exceeded"})
		},
	})

	app.Post("/health/sync", middleware.Authorized(), syncLimiter, postHealthSync)
	app.Post("/health/disconnect", middleware.Authorized(), postHealthDisconnect)
	app.Get("/health/stats", middleware.Authorized(), getHealthStats)
	app.Get("/health/daily", middleware.Authorized(), getHealthDaily)
	app.Get("/health/session/:id", middleware.Authorized(), getHealthSession)
	app.Get("/health/manifest", middleware.Authorized(), getHealthManifest)
	app.Get("/health/readiness/today", middleware.Authorized(), getReadinessToday)
}

func getReadinessToday(c *fiber.Ctx) error {
	loc, err := service.ParseTimezone(c.Get("X-Timezone"))
	if err != nil {
		middleware.Log(c).Warn().Err(err).Msg("invalid timezone header")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("invalid timezone: %v", err)})
	}

	resp, err := service.GetReadinessToday(c.Locals("userID").(uuid.UUID), loc, c.Query("refresh") == "1")
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to get readiness")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get readiness"})
	}
	if resp == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "no recovery data available"})
	}
	return c.JSON(resp)
}

func postHealthSync(c *fiber.Ctx) error {
	var req model.HealthSyncRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// validate X-Timezone header and reject if invalid
	loc, err := service.ParseTimezone(c.Get("X-Timezone"))
	if err != nil {
		middleware.Log(c).Warn().Err(err).Msg("invalid timezone header")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("invalid timezone: %v", err)})
	}

	resp, err := service.SyncHealthData(c.Locals("userID").(uuid.UUID), req, loc)
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to sync health data")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "sync failed"})
	}

	return c.JSON(resp)
}

func getHealthStats(c *fiber.Ctx) error {
	resp, err := service.GetHealthStats(c.Locals("userID").(uuid.UUID))
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to get health stats")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get health stats"})
	}
	return c.JSON(resp)
}

func postHealthDisconnect(c *fiber.Ctx) error {
	if err := service.DisconnectHealth(c.Locals("userID").(uuid.UUID)); err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to disconnect health data")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "disconnect failed"})
	}

	return c.SendStatus(http.StatusOK)
}

func getHealthDaily(c *fiber.Ctx) error {
	// validate X-Timezone header
	loc, err := service.ParseTimezone(c.Get("X-Timezone"))
	if err != nil {
		middleware.Log(c).Warn().Err(err).Msg("invalid timezone header")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("invalid timezone: %v", err)})
	}

	resp, err := service.GetHealthDaily(c.Locals("userID").(uuid.UUID), loc)
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to get health daily")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get health daily"})
	}
	return c.JSON(resp)
}

func getHealthSession(c *fiber.Ctx) error {
	trainingID := c.Params("id")
	if trainingID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "training id is required"})
	}

	parsedID, err := uuid.Parse(trainingID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid training id"})
	}

	// verify training belongs to user (owner or partner)
	userID := c.Locals("userID").(uuid.UUID)
	if !service.UserCanAccessTraining(userID, trainingID) {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "training not found"})
	}

	session, err := service.GetExerciseSessionForTraining(parsedID, userID)
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to get health session")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get health session"})
	}

	if session == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "no health session linked"})
	}

	return c.JSON(session)
}

func getHealthManifest(c *fiber.Ctx) error {
	resp, err := service.GetHealthManifest(c.Locals("userID").(uuid.UUID))
	if err != nil {
		middleware.Log(c).Error().Err(err).Msg("failed to get health manifest")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get health manifest"})
	}
	return c.JSON(resp)
}
