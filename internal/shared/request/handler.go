package request

import (
	"errors"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"

	"github.com/gofiber/fiber/v3"
)

func HandleValidationError(c fiber.Ctx, err error) error {
	switch e := err.(type) {
	case *ValidationError:
		return response.Error(c, fiber.StatusUnprocessableEntity, "validation error", e.Errors)

	default:
		switch {
		case errors.Is(err, ErrInvalidRequestBody):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, ErrInvalidPathParams):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, ErrInvalidRequestQuery):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		default:
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
		}
	}
}
