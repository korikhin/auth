package response

import "strings"

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

func Error(msg string, d ...string) Response {
	r := Response{
		Status:  StatusError,
		Message: msg,
	}

	if len(d) > 0 {
		r.Details = strings.Join(d, ", ")
	}
	return r
}
