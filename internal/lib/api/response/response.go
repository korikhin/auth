package response

import (
	"fmt"
	"strings"
)

// Statuses
const (
	StatusOK    = "ok"
	StatusError = "error"
)

const (
	detailsMaxLength = 256
	detailsEtc       = " [...]"
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
		Status:  StatusOK,
		Message: msg,
	}
}

func Error(msg string, details ...any) Response {
	r := Response{
		Status:  StatusError,
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
