package health

import (
	"net/http"

	"github.com/studopolis/auth-server/internal/lib/api/response"
	"github.com/studopolis/auth-server/internal/lib/http/codec"
)

func New() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		codec.JSONResponse(w, r, response.Ok("", http.StatusOK))
	}

	return http.HandlerFunc(handler)
}
