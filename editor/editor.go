// Visual editor for HTML and Go templates, utilizing Bootstrap.
package editor

// need something to provide a list of pages
// and a way to read/writ them

// need something to provide a list of page templates (simple, impl tool to iterate over a dir would do)

// need something to provide our list of templates that can be included and long with the tooling to
// configure them;
// one way to do it is to have an endpoint that knows how to take the arguments from a template call
// (hm, or actually the full entire tag itself - it might be beneficial to have other arbitrary bits
// of go tempalting be able to be inserted/edited, etc. this would mean that this is wrong:
// <go_template name='/stuff/carousel.gohtml' args='{...}'>
// and it should be more like:
// <go-tag content="{{template &quot;/stuff/carousel.gohtml&quot; Call &quot;{...}&quot;}}">
// )
// and convert it into a configuration screen - that can then be implemented by whatever plugins, possibly
// with just a ServeHTTP call
//
// we probably want one thing to output the form for editing, and another for the preview - but both
// would probably be implemtned by the same module
