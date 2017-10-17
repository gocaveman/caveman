package users

// default user type - best if this type is not exported but that might be taking it too far
// some intefaces that can be used to abstract the user struct from common data needed from it
// like username, roles, email, check password? etc.
// session just stores userid and expiration - use encryption token stuff from webutil
// (come up with a good name "token" may not be the best and it's duplicate of uifiles)
// handler pulls currently logged in user and attaches to context
// login/logout/create endpoints
// user impersonation is done with a separate cookie and a separate handler
// will need subpackage for pages - both admin pages and public login stuff, password reset, etc.
// figure out oauth
