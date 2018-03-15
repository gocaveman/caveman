package webutil

import (
	"fmt"
	"io"
	"strings"
)

// TODO: take a good look at the various go routers out there and find a good pattern for how to do this.
/*

Some links:
https://github.com/julienschmidt/httprouter
https://www.nicolasmerouze.com/guide-routers-golang/
https://husobee.github.io/golang/url-router/2015/06/15/why-do-all-golang-url-routers-suck.html

https://github.com/gorilla/mux/blob/master/regexp.go#L27

*/

// PathParse is a thin wrapper around fmt Sscanf, substituing slashes for spaces.
// The path provided get slashes replaced with strings, then is whitespace trimmed.
// And the result is fed to `fmt.Sscanf`.  An error will be returned if the argument
// types do not match according to Sscanf, or if it could parse too few or
// too many arguments.  The objective is provide a very simple way to rapidly parse
// paths using syntax Go developers are already familiar with.
func PathParse(path string, format string, a ...interface{}) error {

	fpath := strings.TrimSpace(strings.Replace(path, "/", " ", -1))
	fformat := strings.TrimSpace(strings.Replace(format, "/", " ", -1)) + " %s" // additional param so we can detect overruns

	var tmps string
	var args []interface{}
	args = append(args, a...)
	args = append(args, &tmps)

	n, err := fmt.Sscanf(fpath, fformat, args...)

	if n == len(a) && (err == io.EOF || err == nil) {
		return nil
	}

	return err
}

// HasPathPrefix is similar to strings' HasPrefix but intended for paths.
// Matches if path is equal to prefix or if it starts with prefix+"/".
// I.e. prefix "/thepath" will match paths "/thepath", "/thepath/", "/thepath/something/else",
// but not "/thepathogen". Useful for testing "is this path logically the
// same directory as what I'm serving."
func HasPathPrefix(path string, prefix string) bool {
	if path == prefix {
		return true
	}
	// special case to make prefix "/" match everything as would be expected
	if prefix == "/" {
		prefix = ""
	}
	return strings.HasPrefix(path, prefix+"/")
}
