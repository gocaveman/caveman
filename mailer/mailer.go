// Provides a wrapper around email sending including common features like HTML+text
// template rendering and testing tools.
package mailer

// Should expect to be wired with a Renderer, but do not import renderer package unless completely unavoidable.
// One implementation should just send via SMTP.  Need to take into account the common authentication scenarios
// and also make it pluggable.
// Can disable SMTP sending for local dev
// Can enable (separate from SMTP sending) a lot of recent messages for debug purposes, can write to db or
// also just use inmemory sqlite3 if no db present. Ensure it doesn't flood with a huge history of messages,
// probably a DELETE FROM ... WHERE date_created < ? would do it.
// Admin UI can be used to view outbound messages for debugging purposes.  Hm, I just wonder... if it's possible
// to record what the input was during the rendering and then allow the user to refresh this while editing the
// markup - that would be *crazy* but very useful (could be not too bad if we dont' support it on a cluster -
// which is probably fine since it's a debug/dev tool - we just keep a reference to the context that was used
// during the email render and allow the user to "replay" it; we could even try GOBing the input, including
// the context - it's worth a try).  Short of that, it should just be easy to see the messages
// that are being sent.
// It goes without saying that it should be fully usable from tests.
