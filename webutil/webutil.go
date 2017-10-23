// General utilities for web serving.
//
// To qualify for inclusion in this package, the thing in question must be:
//
// Not already solved by or included in the Go stdlib.
//
// Not already solved by or included in a third party package or cost of including that package outweighs the cost of including it here.
//
// Generally useful for web serving.  Features which are intensely Caveman-specific belong elsewhere.  Features which follow a pattern used or introduced in Caveman but don't directly depend on it may be okay.
//
// Don't fit cleanly into one of the other Caveman packages.
//
// Small enough that they don't deserve their own package.
//
// Simple enough to not bloat this package.
//
// Complex enough to bother putting in a separate package at all (avoid one-line functions).
//
// Interfaces which connect disparate packages without introducing dependencies may be okay. (e.g. DataSource, ReadSeekCloser)
//
// Avoid introducing non-stdlib dependencies, pretty much at all costs.
//
// Must be worth maintaining as part of this package (as opposed to e.g. just copying and pasting it around).
//
package webutil

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
)

// ErrNotFound is a generic "not found" error.  Useful to communicate that generic concept
// across packages without introducing dependencies.
var ErrNotFound = errors.New("not found")

// MainOnly checks the call stack to ensure that the caller is in the main package.
// Used to defend against inexperienced developers trying to read from a registry anywhere
// but in the main package.  The argument says how many levels of the stack to remove.
// Use 0 to indicate the function that called this one, 1 to indicate it's caller, etc.
func MainOnly(n int) {

	stackbuf := make([]byte, 4096)
	stackbuf = stackbuf[:runtime.Stack(stackbuf, false)]

	// log.Printf("STACK\n%s", stackbuf)

	// sample stack:

	// goroutine 1 [running]:
	// main.main.func1()
	// 	/Volumes/Files/git/caveman/src/github.com/gocaveman/quickstart-full/main.go:60 +0x77
	// main.main()
	// 	/Volumes/Files/git/caveman/src/github.com/gocaveman/quickstart-full/main.go:62 +0x61d

	r := bufio.NewReader(bytes.NewReader(stackbuf))

	r.ReadString('\n') // "goroutine N ..."

	c := -2
	for {

		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}

		if c < n {
			c++
			continue
		}

		// lines starting with tabs are the filename:lineno ones, skip those
		if strings.HasPrefix(line, "\t") {
			continue
		}

		// first thing before the period should be the package name
		lineParts := strings.SplitN(line, ".", 2)

		if lineParts[0] != "main" {
			panic(fmt.Errorf("call must be made from main package, found %q instead", line))
		}

		break

	}

}
