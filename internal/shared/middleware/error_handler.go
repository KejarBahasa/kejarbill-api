package middleware

import (
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

func ErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return response.Error(c, code, message, nil)
}
