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

func formatErrors(e validator.ValidationErrors) error {
	var msg []string
	for _, err := range e {
		switch err.ActualTag() {
		case "required":
			msg = append(msg, fmt.Sprintf("field %s is required", err.Field()))
		case "email":
			msg = append(msg, fmt.Sprintf("field %s is not a valid email", err.Field()))
		default:
			msg = append(msg, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return errors.New(strings.Join(msg, ", "))
}
