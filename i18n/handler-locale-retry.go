package i18n

func NewLocaleRetryHandler(h interface{}, debug bool) *LocaleRetryHandler {
	return &LocaleRetryHandler{Handler: h, Debug: debug}
}

type LocaleRetryHandler struct {
	Handler interface{}
	Debug   bool
}

// TODO: needs to extract the current locales for the request and try one after the other on
// h, ignoring 404s until one succeeds or the locales are exhausted.
