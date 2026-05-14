package request

import (
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/validator"

	"github.com/gofiber/fiber/v3"
)

func ValidateBody(
	c fiber.Ctx,
	req any,
) error {

	if err := c.Bind().Body(req); err != nil {
		return response.Error(
			c,
			fiber.StatusBadRequest,
			"invalid request body",
			nil,
		)
	}

	if err := validator.Validate.Struct(req); err != nil {
		return response.Error(
			c,
			fiber.StatusBadRequest,
			"validation error",
			validator.ParseValidationError(err),
		)
	}

	return nil
}
