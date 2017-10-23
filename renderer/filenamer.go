package renderer

import (
	"path"
	"strings"
)

// FileNamer converts from a path in a URL to a list of files (in sequence to attempt) that may correspond to it.
// A zero length/nil response is valid and means no attempt should be made to resolve this to a file (usually
// falling through to a 404 or possibly default redirect).
type FileNamer interface {
	FileNames(path string) []string
}

// NewDefaultFileNamer creates a new DefaultFileNamer with the extensions you provide or if none provided uses
// ".gohtml", ".html" and ".md".
func NewDefaultFileNamer(exts ...string) *DefaultFileNamer {
	if len(exts) == 0 {
		exts = []string{".gohtml", ".html", ".md"}
	}
	return &DefaultFileNamer{
		Extensions:    exts,
		DisablePrefix: "_",
	}
}

// DefaultFileNamer provides path mapping based on some sensible defaults.  A list of supported extensions must be provided
// and each filename returned is multiplied by the extensions provided, one for each, in the sequence provided.  The
// basic idea is that for each file, there is only one URL path which maps to it.  Disallowing multiple URL paths that target
// the same file reduces confusion and enforces consistency.  This is not a requirement for all FileNamer implementations,
// but is the default behavior as provided by DefaultFileNamer. Conveniences to the user in the form of redirects that
// fix common path errors by redirecting to the correct path are provided by the pages package.
//
// The rules implemented here are as follows:
//
// Paths ending in a slash are converted to .../index[.ext], e.g "/" -> "/index.gohtml", "/somedir/" -> "/somedir/index.gohtml".
//
// Paths which have "index" as the filename component always return nil, this is to avoid having multiple paths for the same filename.
//
// Paths which already have an extension e.g. "/whatever.html" always return nil.
//
// Paths without an extension return a slice with the Extensions appended, e.g. "/whatever" -> ["/whatever.gohtml", "/whatever.html", "/whatever.md"]
//
// Paths which are "unclean" (path.Clean("/"+p) returns something different, with the exception of the trailing slash)
// will always return nil.
//
// Also, if any path component (folder or file name begins with "_", no filenames will be returned, effectively disabling that path).
// This is intended to provide an easy way to make view templates that are not servable publicly but can still easily be called
// from a handler/controller.
type DefaultFileNamer struct {
	// extensions to try
	Extensions []string
	// path components which start with this prefix are disabled, empty string means none are disable
	DisablePrefix string
}

func (fn *DefaultFileNamer) withExts(p string) []string {
	ret := make([]string, 0, len(fn.Extensions))
	for _, e := range fn.Extensions {
		ret = append(ret, p+e)
	}
	return ret
}

func (fn *DefaultFileNamer) FileNames(p string) []string {

	newp := path.Clean("/" + p)

	// re-append the slash if it was there
	if !strings.HasSuffix(newp, "/") && strings.HasSuffix(p, "/") {
		newp = newp + "/"
	}

	// disallow unclean paths
	if newp != p {
		return nil
	}

	p = newp

	// check for any path component that has the disable prefix, if so then bail
	if fn.DisablePrefix != "" {
		parts := strings.Split(p, "/")
		for _, p := range parts {
			if strings.HasPrefix(p, fn.DisablePrefix) {
				return nil
			}
		}
	}

	pbase := path.Base(p)

	// slashes mean "index"
	if strings.HasSuffix(p, "/") {
		return fn.withExts(p + "index")
	}

	// disallow "index" by name
	if pbase == "index" {
		return nil
	}

	// disallow pages with existing extensions
	if path.Ext(pbase) != "" {
		return nil
	}

	// whatever is left should be attempted with all extensions
	return fn.withExts(p)

}
