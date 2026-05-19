package route

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/user/handler"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"

	"github.com/gofiber/fiber/v3"
)

func UserRoute(
	api fiber.Router,
	authMiddleware *middleware.AuthMiddleware,
	userHandler *handler.UserHandler,
) {

	user := api.Group("/users", authMiddleware.Protected)

	user.Get("/me", userHandler.Me)
}
