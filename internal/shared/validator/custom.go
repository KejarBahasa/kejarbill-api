package validator

import (
	"regexp"
	"strings"
	"unicode"

	goValidator "github.com/go-playground/validator/v10"
)

// regex pattern to reject any string that contains whitespace or non-ascii characters
var asciiRegex = regexp.MustCompile(`^[[:print:]]+$`)

// regex pattern to validate username
// - only contain letters, numbers, underscores, periods
// - cannot start with a period
// - cannot end with a period
var usernameRegex = regexp.MustCompile(`^[a-z0-9_]([a-z0-9_.]*[a-z0-9_])?$`)

func registerCustomValidators() {
	Validate.RegisterValidation("password_strength", func(fl goValidator.FieldLevel) bool {
		password := fl.Field().String()
		runes := []rune(password)

		// min=8, max=128
		if len(runes) < 8 || len(runes) > 128 {
			return false
		}

		if strings.Contains(password, " ") {
			return false
		}

		// reject any string that contains whitespace or non-ascii characters
		if !asciiRegex.MatchString(password) {
			return false
		}

		// must contain 3 of the following types: number, alpha, special chars
		var (
			hasNumber  bool
			hasAlpha   bool
			hasSpecial bool
		)

		for _, char := range runes {
			switch {
			case unicode.IsNumber(char):
				hasNumber = true
			case unicode.IsLetter(char):
				hasAlpha = true
			case unicode.IsPunct(char) || strings.ContainsRune("$&+=<>^`|~", char):
				hasSpecial = true
			}
		}

		return hasNumber && hasAlpha && hasSpecial
	})

	Validate.RegisterValidation("username", func(fl goValidator.FieldLevel) bool {
		// - only contain letters, numbers, underscores, periods
		// - must be between 3 and 50 characters long
		// - cannot start with a period
		// - cannot end with a period
		// - cannot contain consecutive periods

		username := fl.Field().String()

		if len(username) < 3 || len(username) > 50 {
			return false
		}
		if !usernameRegex.MatchString(username) {
			return false
		}

		if strings.Contains(username, "..") {
			return false
		}

		return true
	})
}
