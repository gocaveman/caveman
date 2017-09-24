package caveman

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
)

var ErrNotJSONContent = errors.New("cannot parse, not json content")

func JSONUnmarshalRequest(r *http.Request, v interface{}) error {

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

func JSONMarshalResponse(w http.ResponseWriter, status int, v interface{}) error {

	w.Header().Set("content-type", "application/json")

	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	err := enc.Encode(v)
	if err != nil {
		return err
	}

	return nil
}
