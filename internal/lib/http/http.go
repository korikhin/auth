package http

// Headers
const (
	HeaderAccept          = "Accept"
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderAuth            = "Authorization"
	HeaderAuthorization   = "Authorization"
	HeaderCacheControl    = "Cache-Control"
	HeaderContentLength   = "Content-Length"
	HeaderContentType     = "Content-Type"
	HeaderDNT             = "DNT"
	HeaderIfModifiedSince = "If-Modified-Since"
	HeaderKeepAlive       = "Keep-Alive"
	HeaderOrigin          = "Origin"
	HeaderRange           = "Range"
	HeaderUserAgent       = "User-Agent"

	HeaderCSRFToken     = "X-CSRF-Token"
	HeaderCustomHeader  = "X-CustomHeader"
	HeaderRequestedWith = "X-Requested-With"
	HeaderRequestID     = "X-Request-ID"
	HeaderRequiredRole  = "X-Required-Role"
)

// Content types
const (
	ContentTypeJSON = "application/json"

	// TODO: Ensure that the API complies with RFC 7807
	ContentTypeProblemJSON = "application/problem+json"
)
