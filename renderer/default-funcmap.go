package renderer

import "html/template"

// DEFAULT_FUNCMAP is the FuncMap used if not overridden. (See New() function.)
// Generally, it is not recommended to change the FuncMap.  Prefer assigning
// objects to the context as this is much less likely to cause dependency problems
// and namespace issues.  The default FuncMap is really only here to provide easy
// access to functionality that aides in template rendering in the most general way
// and is common virtualy all applications.
var DEFAULT_FUNCMAP template.FuncMap = template.FuncMap{
	"CallJSON": func(ctx interface{}) interface{} { return ctx }, // FIXME: need to figure this out...
	// TODO: probably HTML, CSS, JS, look around and see if anything else super common
}

// TODO: it probably makes sense to take some common functionality from the Go stdlib and
// expose it to templates with sane naming. For example:
// NewStringsHandler().... results in a "strings" context value which has functions on it
// like HasPrefix(), Split(), Join(), ToLower(), etc.
// FIGURE OUT WHERE THIS SHOULD GO!  I'M NOT SURE THAT renderer IS THE RIGHT PLACE FOR IT,
// ALTHOUGH THERE'S A STRONG ARGUMENT FOR THAT. (MAYBE "renderutil")
