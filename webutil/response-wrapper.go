package webutil

// TODO: a struct (WrapResponseWriter) that cleanly wraps a ResponseWriter
// (can be embedded easily by another struct to add functionality)

// TODO: a struct that embeds WrapResponseWriter to implement context cancellation
// when WriteHeader is called.  Name: CancelingResponseWriter

// TODO: we could also add one here that dumps everything to the log, lower priority
// but probably useful (it should be smart enough to ungzip what GzipResponseWriter
// has done in order to make it human readable).
