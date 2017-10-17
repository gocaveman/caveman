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
