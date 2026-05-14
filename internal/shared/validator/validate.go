package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func ParseValidationError(
	err error,
) map[string]string {

	errors := map[string]string{}

	for _, e := range err.(validator.ValidationErrors) {
		field := strings.ToLower(e.Field())

		switch e.Tag() {

		case "required":
			errors[field] = "field is required"

		case "email":
			errors[field] = "invalid email format"

		case "min":
			errors[field] = "minimum length not met"

		default:
			errors[field] = "invalid value"
		}
	}

	return errors
}
