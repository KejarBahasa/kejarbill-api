package request

import (
	"github.com/KejarBahasa/kejarbill-api/internal/shared/validator"

	"github.com/gofiber/fiber/v3"
)

func ValidateBody(c fiber.Ctx, req any) error {
	if err := c.Bind().Body(req); err != nil {
		return ErrInvalidRequestBody
	}

	if err := validator.Validate.Struct(req); err != nil {
		return &ValidationError{
			Errors: validator.ParseValidationError(req, err),
		}
	}

	return nil
}
