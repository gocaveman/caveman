// Registry for translations provided by various components.
package i18nregistry

import "github.com/gocaveman/caveman/webutil"

func MustRegister(seq float64, name string, t i18n.Translator) {
}

func MakeTranslator(contents webutil.NamedSequence, debug bool) i18n.Translator {
}
