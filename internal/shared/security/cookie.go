package security

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

func SetRefreshCookie(c fiber.Ctx, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/api/auth",
		Expires: time.Now().Add(
			30 * 24 * time.Hour,
		),
	})
}

func ClearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Expires:  time.Now().Add(-time.Hour),
	})
}
