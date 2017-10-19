package dbutil

// something to scan struct fields for common validation - so you can tag your structs with things like
// min and max length, email formatted, whatever.
// helper func makes this a one-liner to implement a Validate() error function on a model object, with
// a return value that can be WriteJSON or whatever and is useful in the browser
// the whole idea is that validation is really just a convention along with a helper func or two
// no depdendency on a particular ORM or that an ORM is used at all

// think about if it makes sense to generate default HTML form elements from this validation data
// - it could be a simple crutch to build a form very rapidly; it's an odd collapse of tiers
//   but it also shouldn't introduce dependnecies or other weird stuff - it's really just a
//   helper

// we're going to need gorm (or sqlx/modl!) tooling for looking up relations and other basic
// record operations in HTML (in a form), otherwise you end up writing a controller that
// has to load all kinds of crap that really isn't controller-specific - it's just simple
// data calls needed by the page.  maybe we should preclude any operations which write! (probably should)

// default "controllers" - FormController, ListController from OctoberCMS - although maybe not, might
// be better to code generate (cmd line tool) simple controllers that have the REST endpoint stuff on
// them and use helper funcs to make them efficient and concise - this is probably better than some
// bit giant class that you extend and configure
// cmd line to generate a bunch of crap for a particular model
