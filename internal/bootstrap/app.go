package bootstrap

import (
	"time"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

func BuildApp(dep *Dependency) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      dep.Config.AppName,
		ErrorHandler: middleware.ErrorHandler,
		ReadTimeout:  10 * time.Second,
	})

	RegisterMiddleware(app, dep.Config.AppEnv)

	RegisterRoute(app, dep)

	return app
}
