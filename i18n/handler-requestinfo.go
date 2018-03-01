package i18n

// hm, struct or interface... maybe struct but T or whatever else can be interface - but again, hm,
// the concept of "give me the locales for this page" really should be an interface...
// type RequestInfo struct {
// 	Locales() []string
// 	Translator() Translator
// 	LocaleTranslator() LocaleTranslator
// 	Processor Processor
// }

// probaby a T call right on here and the template rendering stuff too, although rendering should also
// be avail separatey for Go code

// Group(g string)

// T
// TProc

// T2
// T2Proc

// we also should have a call that splits by colon into group and key - for cases where a string field is
// coming in from something that doesn't support translations or isn't aware of the locale and we need to
// be able to express group:key as one string
// TODO: make sure this makes it into the documentation article.

// func NewRequestInfoHandler(tr Translator) webutil.HandlerChain {

// }

// type RequestInfoHandler struct {
// 	Translator Translator
// }
