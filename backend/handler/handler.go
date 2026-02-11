package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/streambinder/vigor/handler/middleware"
)

// Init initializes the Fiber application with middleware and routes.
// This function must be called before starting the server.
func Init() *fiber.App {
	app := fiber.New()

	// Register route handlers with no middleware
	initHealth(app)

	// Register request ID middleware
	app.Use(requestid.New())

	// Register compression middleware
	app.Use(compress.New())

	// Register CORS middleware
	app.Use(middleware.CORS())

	// Register route handlers with CORS
	initProxy(app)

	// Register logging middleware
	app.Use(middleware.Logging())

	// Register route handlers with logging and CORS
	initAvatar(app)
	initSession(app)
	initOauth(app)
	initUser(app)
	initGym(app)
	initEquipment(app)
	initGoal(app)
	initMuscle(app)
	initMovementFamily(app)
	initMethodology(app)
	initTraining(app)
	initActivity(app)
	initProgress(app)
	initReport(app)

	return app
}
