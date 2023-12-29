package cors

import (
	"net/http"

	"github.com/studopolis/auth-server/internal/config"
	httplib "github.com/studopolis/auth-server/internal/lib/http"

	"github.com/gorilla/handlers"
)

func New(config config.CORS) func(http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins(config.AllowedOrigins),
		handlers.AllowedMethods([]string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
		}),
		handlers.AllowedHeaders([]string{
			httplib.HeaderAuth,
			httplib.HeaderCacheControl,
			httplib.HeaderContentType,
			httplib.HeaderDNT,
			httplib.HeaderIfModifiedSince,
			httplib.HeaderKeepAlive,
			httplib.HeaderRange,
			httplib.HeaderRequestedWith,
			httplib.HeaderUserAgent,
		}),
		handlers.ExposedHeaders([]string{
			httplib.HeaderRequestID,
		}),
		handlers.AllowCredentials(),
		handlers.MaxAge(config.MaxAge),
	)
}
