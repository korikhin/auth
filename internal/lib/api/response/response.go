package response

import (
	"fmt"
	"strings"
)

const (
	StatusOK    = "OK"
	StatusError = "Error"
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
		m := []string{}
		for _, d := range details {
			m = append(m, fmt.Sprint(d))
		}
		r.Details = strings.Join(m, ", ")
	}

	return r
}

func InternalError() Response {
	return Error("Internal service error")
}
