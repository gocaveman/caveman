package handlerregistry

import (
	"net/http"

	"github.com/gocaveman/caveman/webutil"
)

func RegisterHandler(seq float64, name string, h http.Handler) {

}

func RegisterChainHandler(seq float64, name string, ch webutil.ChainHandler) {

}
