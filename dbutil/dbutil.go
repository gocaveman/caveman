// Database/model utilities.
package dbutil

import "github.com/gocaveman/caveman/webutil"

// TODO: Record locking is another pain-in-the-ass but common pattern that might be good to have a tool for.
// Queues for example or tasks that can only be run on one server - you perform the lock operation atomically,
// typicaly by putting a timestamp in an otherwise null field indicating the lock time and if your update worked
// then you're good.  Need an unlock mechanism, an update as well in case it's a long-running thing;
// And something needs to clean up old stale locks.  Was thinking
// this would need to integrate with ORM tooling but realistically a database/sql solutin would probably work
// just fine even when an ORM is in use.

var ErrNotFound = webutil.ErrNotFound
