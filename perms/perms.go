package perms

// figure out if structs or interfaces
// func HasPerm(roles []string, perm string) bool - however we do it this the main thing that needs to be implemented
// people will use it by getting the role list off of a user and calling this to check a perm
// registry is important for this one - probably permregistry subpackage - and components can register their permissions
// with the different roles, and in main.go that can all be discarded or edited or whatever
// could be a db implementation but not priority for v1; make sure the structure looks good
