package gen

// listing - probably loads ajax (rather than page data), because
// search will need that anyway

// mobile-first

// embedability - think about what happens if we want to move this to an include file
//  and call it from JS in a modal or something, what can we do to make that scenario painless

// sort by columns

// max record limit

// ability to add the idea that you can drag and drop a sequnece for things - not core functionalityu
// but needs to be addable if needed; possibly a separate generator

// paging - think about how it integrates with search, and how permalinks work

// fancy search: Here’s the idea I’m leaning toward:
// In our listing pages, we basically have a pattern where we
// have a set of which fields are searched.  The case of a single
// ‘word’ typed in (i.e. no space) to the search, would result in
// an AJAX query for each of the fields, in sequence, with a max
// limit on each one, for example 50.  It also stops when it gets
// 50 results.  Example - we search “email”, “phone”, “company_name”,
// “first_name”, “last_name” - in that sequence.  And just fill up to
// 50 records.  So the guy types in a phone number, great, one match.
// Same with and email.  But if he types in “joe” it will end up
// searching for “joe%” in those fields and probably return some
// things for the company, first and last. It would be a bit slower,
// but it would be really cool and useful.  And once the pattern is there,
// pretty easy to do. (And if you want fast and simple single-field
// searches, you just make the field list just one field and it has
// the right behavior.)

// think through how case-sensitivity works
// current idea: both mysql and sqlite3 provide case-insensitive collations, it
// makes good sense to just use the right collation for fields in the migrations
// we provide and then like 'joe%' will do the right thing.  (need to make sure that
// string keys (base64 uuids) are case sensitive but pretty much everything else is
// case insensitive); for postgress we need to use ILIKE, maybe it can be a hack
// in the store layer to structure the query in that way (or look at adding it to
// dbrobj to help abstract it away) - either way, get the database to do it for us -
// it's somewhat obscure but is actually supported if you know what you're doing;
// user can also just change the collation for a field if case-insensitivity is wrong,
// but for common cases it will be correct.

// toggle which fields it searches?

// option to edit records inline! (to add them too? delete? - what about field validation...
// this starts to get really compilcated, see if it's worth it, or if there is some
// simplification than can be done to limit the complexity on how this is implemented -
// it would very very cool though...)  See if there is a way to combine logic with the
// detail page, otherwise it may start to be duplicative... such as...:
// error message handling on saving
// including login redirection for when login times out
