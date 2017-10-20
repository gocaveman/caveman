// General utilities for web serving.
// To qualify for inclusion in this package, the thing in question must be:
// <ul>
// <li>Not already solved by or included in the Go stdlib.
// <li>Not already solved by or included in a third party package or cost of including that package outweighs the cost of including it here.
// <li>Generally useful for web serving.  Features which are intensely Caveman-specific belong elsewhere.  Features which follow a pattern used or introduced in Caveman but don't directly depend on it may be okay.
// <li>Don't fit cleanly into one of the other Caveman packages.
// <li>Small enough that they don't deserve their own package.
// <li>Simple enough to not bloat this package.
// <li>Complex enough to bother putting in a separate package at all (no one-line functions).
// <li>Interfaces which connect disparate packages without introducing dependencies may be okay. (e.g. DataSource, ReadSeekCloser)
// <li>Avoid introducing non-stdlib dependencies, pretty much at all costs.
// <li>Must be worth maintaining as part of this package (as opposed to e.g. just copying and pasting it around).
// </ul>
package webutil

import (
	"errors"
)

// ErrNotFound is a generic "not found" error.  Useful to communicate that generic concept
// across packages without introducing dependencies.
var ErrNotFound = errors.New("not found")

// MainOnly checks the call stack to ensure that the caller is in the main package.
// Used to defend against inexperienced developers trying to read from a registry anywhere
// but in the main package.  The argument says how many
func MainOnly(n int) {

	// runtime.Stack(buf, all)
	// TODO: implement, not sure if this is actually possible, but if it is... do it.

}
