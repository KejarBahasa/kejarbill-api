package route

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/handler"

	"github.com/gofiber/fiber/v3"
)

func AuthRoute(
	api fiber.Router,
	authHandler *handler.AuthHandler,
) {
	auth := api.Group("/auth")

	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh-token", authHandler.RefreshToken)
	auth.Post("/logout", authHandler.Logout)
}
