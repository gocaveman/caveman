// Tooling to take flat files and synchronize them with a sqlite database for easy access use of
// large amounts of static data, generally versioned with the website project.
package file2sqlite

// developer provides a struct, can be unmarshaled from yaml or json or other
// struct also maps to db columns - using sqlx
// one way is one file per record
// another is one file with an array with a bunch of records
// either way it loads into in mrmory or tempfile sqlite db and then becomes
// queryable with sqlx
// the idea is it gets run at startup and you can make flat files into db
// tables easily
// optional writing allows writes to update sqlite and also update backing file
// - see if there's a way to do this transparently but don't die over it -
//   think through how this would be used with a controller; maybe deserving
//   of some template generation stuff
// think about how this works with the page index - will scan view files
// reading the meta blocks and index them in a table like this - will need
// to be able to customize that; and then use the same approach as if
// things were fully database driven - this means that we probably don't
// want to mix and match ORMs - using the same one here will help us a lot;
// or possibly this package works completely with database/sql and/or sqlx
// but then if someone uses a gorm object on top of it that will work just
// fine - that could work well too
