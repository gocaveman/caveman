package pages

// TODO: if pages now knows the logical URL paths of all the pages on the site, we can do some simple redirects
// to avoid the user being tripped up when entering URLs.  examples:
// "/index" -> "/"
// "/somefile.html" -> "/somefile"
// "/somefile" (but it doesn't exist) -> "/somefile/" (where it does exist)
// likewise "/somefile/" -> "/somefile" where that's the valid case
// AHA! We may not need to know the list of pages for this - this could be a more generic thing
// maybe in webutil if the logic is simply "if it's a GET request, try it the other
// way and if that returns something then redirect" - or possibly a FileSystem test (hm, but
// then maybe it would need a copy of the FileNamer being used by the renderer - think this through); plus
// the /index -> / logic.  That would make it super generic and simple but still with the
// intended functionality.
