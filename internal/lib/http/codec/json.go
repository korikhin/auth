package codec

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	ctxlib "github.com/korikhin/auth/internal/lib/context"
	httplib "github.com/korikhin/auth/internal/lib/http"
)

func JSONResponse(w http.ResponseWriter, r *http.Request, v interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(httplib.HeaderContentType, httplib.ContentTypeJSON)
	if status, ok := r.Context().Value(ctxlib.StatusKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write(buf.Bytes())
}

func DecodeJSON(r io.Reader, v interface{}) error {
	defer io.Copy(io.Discard, r)
	return json.NewDecoder(r).Decode(v)
}
