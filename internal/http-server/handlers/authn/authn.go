package authn

import (
	"net/http"

	"github.com/korikhin/auth/internal/lib/api/response"
	"github.com/korikhin/auth/internal/lib/http/codec"
)

func New() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		codec.JSONResponse(w, r, response.Ok("", http.StatusOK))
	}

	return http.HandlerFunc(handler)
}