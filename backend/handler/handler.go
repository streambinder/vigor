package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/streambinder/vigor/handler/middleware"
)


// Init initializes the Fiber application with middleware and routes.
// This function must be called before starting the server.
func Init() *fiber.App {
	app := fiber.New()

	// Register global middleware (order matters!)
	app.Use(middleware.CORS())
	app.Use(middleware.Logging())

	// Register route handlers
	initHealth(app)
	initSession(app)
	initOauth(app)
	initUser(app)
	initGym(app)
	initTraining(app)

	return app
}
