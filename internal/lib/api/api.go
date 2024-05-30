package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	EmptyRequest  = Error("request body is empty")
	InternalError = Error("internal server error")
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

func Ok(msg string) Response {
	return Response{
		Status:  "ok",
		Message: msg,
	}
}

func Error(msg string, details ...any) Response {
	const detailsMaxLength = 255
	const detailsEtc = " [...]"

	r := Response{
		Status:  "error",
		Message: msg,
	}

	if len(details) > 0 {
		var builder strings.Builder
		for i, d := range details {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprint(d))
			if builder.Len() > detailsMaxLength {
				break
			}
		}

		detailsJoined := builder.String()
		if len(detailsJoined) > detailsMaxLength {
			detailsJoined = fmt.Sprintf(
				"%s%s",
				detailsJoined[:detailsMaxLength-len(detailsEtc)], detailsEtc,
			)
		}
		r.Details = detailsJoined
	}

	return r
}

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
		switch f := err.Field(); err.ActualTag() {
		case "required":
			message = fmt.Sprintf("field %s is required", f)
		case "email":
			message = fmt.Sprintf("field %s is not a valid email", f)
		default:
			message = fmt.Sprintf("field %s is not valid", f)
		}
		messages = append(messages, message)
	}

	return errors.New(strings.Join(messages, ", "))
}
