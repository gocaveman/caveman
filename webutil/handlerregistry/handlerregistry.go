// Registry for handlers to be automatically made available in the handler list.
package handlerregistry

import (
	"net/http"

	"github.com/gocaveman/caveman/webutil"
)

func MustRegister(seq float64, name string, h interface{}) {

}

func MustRegisterHandler(seq float64, name string, h http.Handler) {

}

func MustRegisterChainHandler(seq float64, name string, ch webutil.ChainHandler) {

}
