// Registry for handlers to be automatically made available in the handler list.
package handlerregistry

import (
	"net/http"

	"github.com/gocaveman/caveman/webutil"
)

const (
	SeqFirst      = float64(0)   // processed first
	SeqSetup      = float64(10)  // before any middleware is run
	SeqMiddleware = float64(20)  // middleware can modify before controllers are processed
	SeqCtrl       = float64(50)  // controllers do the main HTTP handling
	SeqRender     = float64(80)  // page rendering happens after other stuff is processed
	SeqLast       = float64(100) // processed last
)

var OnlyReadableFromMain = true

var reg webutil.NamedSequence

// Contents returns the current contents of the registry as a NamedSequence.
func Contents() webutil.NamedSequence {
	if OnlyReadableFromMain {
		webutil.MainOnly(1)
	}
	return reg.SortedCopy()
}

func MustRegister(seq float64, name string, h interface{}) {
	reg = append(reg, webutil.NamedSequenceItem{Sequence: seq, Name: name, Value: h})
}

func MustRegisterHandler(seq float64, name string, h http.Handler) {
	MustRegister(seq, name, h)
}

func MustRegisterChainHandler(seq float64, name string, ch webutil.ChainHandler) {
	MustRegister(seq, name, ch)
}
