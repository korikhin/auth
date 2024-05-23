package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Credentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func Validate(c *Credentials) error {
	if err := validator.New().Struct(c); err != nil {
		errs := err.(validator.ValidationErrors)
		return formatErrors(errs)
	}

	return nil
}

func formatErrors(validationErrors validator.ValidationErrors) error {
	var messages []string
	for _, err := range validationErrors {
		var message string
		switch field := err.Field(); err.ActualTag() {
		case "required":
			message = fmt.Sprintf("field %s is required", field)
		case "email":
			message = fmt.Sprintf("field %s is not a valid email", field)
		default:
			message = fmt.Sprintf("field %s is not valid", field)
		}
		messages = append(messages, message)
	}

	return errors.New(strings.Join(messages, ", "))
}
