package request

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	httplib "github.com/studopolis/auth-server/internal/lib/http"
)

var prefix string
var reqID uint64

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}

	var buf [12]byte
	var b64 string

	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}

func RequestID(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(httplib.RequestIDHeader)

		if requestID == "" {
			myID := atomic.AddUint64(&reqID, 1)
			requestID = fmt.Sprintf("%s-%06d", prefix, myID)
		}

		ctx = context.WithValue(ctx, httplib.RequestCtxKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(handler)
}

func GetID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if id, ok := ctx.Value(httplib.RequestCtxKey).(string); ok {
		return id
	}

	return ""
}

func NextID() uint64 {
	return atomic.AddUint64(&reqID, 1)
}
