package gen

// generates form from model

// mobile-first

// embedability - think about what happens if we want to move this to an include file
//  and call it from JS in a modal or something, what can we do to make that scenario painless
//  Vue components look promising, but need to see how to bring in the templates without
//  duplication - possibly a {{define "once /path-to-include.gohtml"}}{{end}} or something
//  is needed to make this work painlessly - that or we give templates a way of deduplicating
//  themselves...  actually that could be really simple, {{if not (.Included "/template-name.gohtml")}}
//  {{.MarkIncluded "/template-name.gohtml"}}, or if we can somehow detect the current template
//  name, we can even do {{if .Once}} ... {{end}} - which would be rad, although maybe
//  {{if .Once "/template-name.gohtml"}} ... {{end}} is more practical.

// can customize fields

// loads page data from controller

// need to be able to pass in something that allows us to look up relations - definitely needs think-through

// error message handling on saving
// including login redirection for when login times out
