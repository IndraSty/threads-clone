package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			fieldName := strings.ToLower(fieldError.Field())
			errors[fieldName] = getErrorMessage(fieldError)
		}
	}

	return errors
}

func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}
