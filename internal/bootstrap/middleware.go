package bootstrap

import (
	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func RegisterMiddleware(app *fiber.App, appEnv string) {
	app.Use(
		recover.New(
			recover.Config{
				EnableStackTrace: appEnv != "production",
				StackTraceHandler: func(c fiber.Ctx, e any) {
					ctx := c.Context()
					log := middleware.LoggerFromContext(ctx)
					log.Error().Any("panic", e).Msg("panic recovered")
				},
			},
		),
	)

	app.Use(requestid.New())

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

	app.Use(middleware.ContextBridge())

	app.Use(middleware.ZeroLog())
}
