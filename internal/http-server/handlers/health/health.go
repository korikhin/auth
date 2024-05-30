package health

import (
	"net/http"

	response "github.com/korikhin/auth/internal/lib/api"
	"github.com/korikhin/auth/internal/lib/http/codec"
)

func New() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		codec.ResponseJSON(w, response.Ok(""), http.StatusOK)
	}

	return http.HandlerFunc(handler)
}
