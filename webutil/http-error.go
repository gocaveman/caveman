package webutil

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

// HTTPError reports an error - does not expose anything to the outside
// world except a unique ID, which can be matched up with the appropriate log
// statement which has the details.
func HTTPError(w http.ResponseWriter, r *http.Request, err error, publicMessage string, code int) error {

	if err == nil {
		err = errors.New(publicMessage)
	}

	id := fmt.Sprintf("%x", time.Now().Unix()^rand.Int63())

	_, file, line, _ := runtime.Caller(1)

	w.Header().Set("x-error-id", id) // make a way for the client to programatically extract the error id
	http.Error(w, fmt.Sprintf("Error serving request (id=%q) %s", id, publicMessage), code)

	log.Printf("HTTPError: (id=%q) %s:%v | %v", id, file, line, err)

	return err
}
