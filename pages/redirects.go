package pages

// TODO: if pages now knows the logical URL paths of all the pages on the site, we can do some simple redirects
// to avoid the user being tripped up when entering URLs.  examples:
// "/index" -> "/"
// "/somefile.html" -> "/somefile"
// "/somefile" (but it doesn't exist) -> "/somefile/" (where it does exist)
// likewise "/somefile/" -> "/somefile" where that's the valid case
