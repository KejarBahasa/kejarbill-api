package validator

import (
	"reflect"
	"strings"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/constants"
	goValidator "github.com/go-playground/validator/v10"
)

var Validate = goValidator.New()

func ParseValidationError(req any, err error, tags ...string) []ValidationError {
	errors := make([]ValidationError, 0)
	validationErrors := err.(goValidator.ValidationErrors)
	tag := constants.RequestTagJSON
	if len(tags) > 0 && tags[0] != "" {
		tag = tags[0]
	}

	for _, e := range validationErrors {
		errors = append(errors, ValidationError{
			Field:   getNamespaceField(req, e.Namespace(), tag),
			Message: parseMessage(e),
		})
	}

	return errors
}

func getNamespaceField(req any, namespace string, tag string) string {
	parts := strings.Split(namespace, ".")

	if len(parts) <= 1 {
		return strings.ToLower(namespace)
	}

	t := reflect.TypeOf(req)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	result := make([]string, 0)

	// SKIP ROOT STRUCT
	currentType := t
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		fieldName := part
		index := ""

		// Handle Index Array/Slice
		if strings.Contains(part, "[") {
			start := strings.Index(part, "[")
			index = part[start:]
			fieldName = part[:start]
		}

		// Get Field
		field, ok := currentType.FieldByName(fieldName)
		if !ok {
			result = append(result, strings.ToLower(fieldName)+index)
			continue
		}

		fieldTag := field.Tag.Get(tag)
		fieldNameMapped := strings.Split(fieldTag, ",")[0]
		if fieldNameMapped == "" || fieldNameMapped == "-" {
			fieldNameMapped = strings.ToLower(fieldName)
		}

		result = append(result, fieldNameMapped+index)

		// Next Type
		fieldType := field.Type

		// Slice
		if fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
		}

		// Pointer
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		currentType = fieldType
	}

	return strings.Join(result, ".")
}

func parseMessage(e goValidator.FieldError) string {
	param := e.Param()

	switch e.Tag() {

	case "required":
		return "field is required"

	case "email":
		return "invalid email format"

	case "uuid":
		return "invalid uuid format"

	case "min":
		if e.Kind().String() == "string" {
			return "minimum length is " + param
		}

		return "minimum value is " + param

	case "max":
		if e.Kind().String() == "string" {
			return "maximum length is " + param
		}

		return "maximum value is " + param

	case "gt":
		return "must be greater than " + param

	case "lt":
		return "must be less than " + param

	default:
		return "invalid value"
	}
}
