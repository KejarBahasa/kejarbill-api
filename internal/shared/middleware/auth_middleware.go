package middleware

import (
	"strings"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

func Protected(pasetoMaker *security.PasetoMaker) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return fiber.ErrUnauthorized
		}

		splitToken := strings.Split(authHeader, " ")

		if len(splitToken) != 2 {
			return fiber.ErrUnauthorized
		}

		token := splitToken[1]

		payload, err := pasetoMaker.VerifyToken(token)
		if err != nil {
			return fiber.ErrUnauthorized
		}

		c.Locals(security.ContextUserID, payload.UserID)

		return c.Next()
	}
}
