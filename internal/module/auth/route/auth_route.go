package route

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/handler"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/middleware"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

func AuthRoute(
	api fiber.Router,
	authHandler *handler.AuthHandler,
	pasetoMaker *security.PasetoMaker,
) {
	auth := api.Group("/auth")

	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh-token", authHandler.RefreshToken)

	auth.Get("/me", middleware.Protected(pasetoMaker), authHandler.Me)
}
