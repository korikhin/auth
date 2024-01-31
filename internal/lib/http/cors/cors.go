package cors

import (
	"net/http"

	"github.com/studopolis/auth-server/internal/config"
	httplib "github.com/studopolis/auth-server/internal/lib/http"

	"github.com/gorilla/handlers"
)

func New(c config.CORS) func(http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins(c.AllowedOrigins),
		handlers.AllowedMethods([]string{
			http.MethodOptions,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
		}),
		handlers.AllowedHeaders([]string{
			httplib.HeaderAccept,
			httplib.HeaderAcceptEncoding,
			httplib.HeaderAuth,
			httplib.HeaderAuthorization,
			httplib.HeaderCacheControl,
			httplib.HeaderContentLength,
			httplib.HeaderContentType,
			httplib.HeaderCustomHeader,
			httplib.HeaderDNT,
			httplib.HeaderIfModifiedSince,
			httplib.HeaderKeepAlive,
			httplib.HeaderOrigin,
			httplib.HeaderRange,
			httplib.HeaderRequestedWith,
			httplib.HeaderRequiredRole,
			httplib.HeaderUserAgent,
			httplib.HeaderCSRFToken,
		}),
		handlers.ExposedHeaders([]string{
			httplib.HeaderRequestID,
		}),
		handlers.AllowCredentials(),
		handlers.MaxAge(c.MaxAge),
	)
}
