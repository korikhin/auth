package codec

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	httplib "github.com/korikhin/auth/internal/lib/http"
)

func JSONResponse(w http.ResponseWriter, v interface{}, statusCode int) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(httplib.HeaderContentType, httplib.ContentTypeJSON)
	w.WriteHeader(statusCode)
	w.Write(buf.Bytes())
}

func DecodeJSON(r io.Reader, v interface{}) error {
	defer io.Copy(io.Discard, r)
	return json.NewDecoder(r).Decode(v)
}
