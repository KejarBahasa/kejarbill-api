package request

import (
	"github.com/KejarBahasa/kejarbill-api/internal/shared/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/validator"

	"github.com/gofiber/fiber/v3"
)

func ValidateQuery(c fiber.Ctx, req any) error {
	if err := c.Bind().Query(req); err != nil {
		return ErrInvalidRequestQuery
	}

	if err := validator.Validate.Struct(req); err != nil {
		return &ValidationError{
			Errors: validator.ParseValidationError(req, err, constants.RequestTagQuery),
		}
	}

	return nil
}
