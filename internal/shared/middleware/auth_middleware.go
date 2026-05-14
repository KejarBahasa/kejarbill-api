package middleware

import (
	"strings"

	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/repository"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type AuthMiddleware struct {
	pasetoMaker *security.PasetoMaker
	authRepo    *repository.AuthRepository
}

func NewAuthMiddleware(pasetoMaker *security.PasetoMaker, authRepo *repository.AuthRepository) *AuthMiddleware {
	return &AuthMiddleware{
		pasetoMaker: pasetoMaker,
		authRepo:    authRepo,
	}
}

func (m *AuthMiddleware) Protected(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return fiber.ErrUnauthorized
	}

	splitToken := strings.Split(authHeader, " ")

	if len(splitToken) != 2 {
		return fiber.ErrUnauthorized
	}

	token := splitToken[1]

	payload, err := m.pasetoMaker.VerifyToken(token)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	if payload.TokenType != security.TokenTypeAccess {
		return fiber.ErrUnauthorized
	}

	user, err := m.authRepo.FindByID(c.Context(), payload.UserID)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	if user.TokenVersion != payload.TokenVersion {
		return fiber.ErrUnauthorized
	}

	c.Locals(security.ContextUserID, payload.UserID)

	return c.Next()
}
