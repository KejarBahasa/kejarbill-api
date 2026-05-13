package route

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/user/handler"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

func UserRoute(
	api fiber.Router,
	userHandler *handler.UserHandler,
	pasetoMaker *security.PasetoMaker,
) {

	user := api.Group("/users", middleware.Protected(pasetoMaker))

	user.Get("/me", userHandler.Me)
}
