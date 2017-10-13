package uifiles

import (
	"net/http"
	"time"
)

type WrongContentRetry struct {
	BlockTime            time.Duration // how long to block before returning
	Num                  int           // how many times do we retry before giving up
	RetryExceededHandler http.Handler  // optional: call this handler after we retry Num times and get called again
}

func (h *WrongContentRetry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: make sure to test this
	panic("not implemented yet")
}

type WrongContentError struct{}

func (h *WrongContentError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "refusing to return incorrect content", 500)
}
