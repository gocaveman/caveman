package i18n

// flat file -> sqlite would be an excellent choice here (although it is actually simpler and
// complex queries are not required, so... we'll see - possibly in the case of large datasets,
// but then the build time would be bad.  yeah, simple in memory cache with LRU or something
// if it gets too big - at least the possibility of plugging that in - probably more the way to go)

// probably want to provide something in the context that can know what the current page's
// locale is, with defaults, and a way to override

// maybe a registry mechanism so other things can provide translations

// at least think through what happens if they want to have one page per translation
// instead of using string lookups, we should facilitate that (although doing it
// with the metadata maybe tricky - but give that some thought too);dont' need to die
// over it but it needs to be feasible if desired.
