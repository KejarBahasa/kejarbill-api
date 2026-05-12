package bootstrap

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func RegisterMiddleware(app *fiber.App) {
	app.Use(recover.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:9000",
		},
		AllowCredentials: true,
		AllowMethods: []string{
			"GET",
			"POST",
		},
	}))
}
