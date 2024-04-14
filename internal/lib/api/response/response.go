package response

import (
	"fmt"
	"net/http"
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
	EmptyRequest  = Error("request body is empty", http.StatusBadRequest)
	InternalError = Error("internal server error", http.StatusInternalServerError)
)

type Response struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

func Ok(msg string, code int) Response {
	if code < 200 || code > 299 {
		code = http.StatusOK
	}

	return Response{
		Code:    code,
		Status:  StatusOK,
		Message: msg,
	}
}

func Error(msg string, code int, details ...any) Response {
	if code < 400 || code > 599 {
		code = http.StatusInternalServerError
	}
	r := Response{
		Code:    code,
		Status:  StatusError,
		Message: msg,
	}

	if len(details) > 0 {
		m := []string{}
		for _, d := range details {
			m = append(m, fmt.Sprint(d))
		}

		detailsJoined := strings.Join(m, ", ")
		// Trim details
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
