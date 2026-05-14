package bootstrap

import (
	authRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/auth/route"
	userRoutePkg "github.com/KejarBahasa/kejarbill-api/internal/module/user/route"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

func RegisterRoute(
	app *fiber.App,
	dep *Dependency,
) {
	app.Get("/", func(c fiber.Ctx) error {
		return response.Success(c, "KejarBill API is running", nil)
	})

	api := app.Group("/api")
	apiV1 := api.Group("/v1")

	// AUTH
	authRoutePkg.AuthRoute(apiV1, dep.AuthHandler)

	// USER
	userRoutePkg.UserRoute(apiV1, dep.UserHandler, dep.AuthMiddleware)
}
