package authenticate

import (
	"net/http"
)

func New() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	return http.HandlerFunc(handler)
}
