package codec

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	httplib "github.com/studopolis/auth-server/internal/lib/http"
)

func JSONResponse(w http.ResponseWriter, r *http.Request, v interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if status, ok := r.Context().Value(httplib.StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write(buf.Bytes())
}

func DecodeJSON(r io.Reader, v interface{}) error {
	defer io.Copy(io.Discard, r)
	return json.NewDecoder(r).Decode(v)
}