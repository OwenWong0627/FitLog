package api

import (
	"database/sql"
	"goapp/packages/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func WithDB(fn func(c *fiber.Ctx, db *sql.DB) error, db *sql.DB) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return fn(c, db)
	}
}

func (a *App) httpServer(db *sql.DB) *fiber.App {
	// Set up HTTP server with fiber
	app := fiber.New()
	app.Use(logger.New())
	app.Use(requestid.New())

	// Set up CORS Middleware management
	api := app.Group("/api")
	api.Use(cors.New(cors.Config{
		AllowOrigins:     config.Config[config.CLIENT_URL] + ", " + config.Config[config.APi_NINJA_URL],
		AllowCredentials: true,
		AllowHeaders:     "Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin",
		AllowMethods:     "POST, OPTIONS, GET, PUT",
		ExposeHeaders:    "Set-Cookie",
	}))

	api.Post("/exercises", a.Exercises)

	api.Post("/login", WithDB(a.Login, db))
	api.Post("/register", WithDB(a.CreateUser, db))
	api.Post("/otp", a.OTP)
	api.Get("/logout", a.Logout)

	// authed routes
	api.Get("/workouts", WithDB(a.Workouts, db))

	return app
}
