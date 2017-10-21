package webutil

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
)

var ErrNotJSONContent = errors.New("cannot parse, not json content")

// ParseJSON parses a request with content-type application/json into the struct you provide.
// Uses json.Decoder.
func ParseJSON(r *http.Request, v interface{}) error {

	ct, _, _ := mime.ParseMediaType(r.Header.Get("content-type"))
	if ct != "application/json" {
		return ErrNotJSONContent
	}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(v)
	if err != nil {
		return err
	}

	return nil
}

// WriteJSON writes a response as content-type application/json.
// Uses json.Encoder.
func WriteJSON(w http.ResponseWriter, v interface{}, status int) error {

	w.Header().Set("content-type", "application/json")

	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	err := enc.Encode(v)
	if err != nil {
		return err
	}

	return nil
}
