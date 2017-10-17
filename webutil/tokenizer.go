package webutil

// TODO: make something generic that addresses this issue of taking arbitrary data and encrypting it for
// URL, cookie/session, etc. The step of JSON marshaling shoudl be simple but not necessarily included,
// people may want to customize how the marshaling is done - but it should be a one-liner.
// Steal from uifiles/token.go as needed, the gzipping part is kinda cool.
// Probably need to have the thing you put on the context to decode it inside a template - you know some
// clown is going to want that like a teenage girl wants Justin Bieber.
