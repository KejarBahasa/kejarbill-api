package request

import (
	"github.com/KejarBahasa/kejarbill-api/internal/shared/validator"

	"github.com/gofiber/fiber/v3"
)

func ValidateParams(c fiber.Ctx, req any) error {
	if err := c.Bind().URI(req); err != nil {
		return ErrInvalidRequestParams
	}

	if err := validator.Validate.Struct(req); err != nil {
		return &ValidationError{
			Errors: validator.ParseValidationError(req, err, "params"),
		}
	}

	return nil
}
