// The uifiles package provides a runtime solution for CSS and JS file minication, combination and inclusion.
// The intention is to obviate the need for tools like webpack which operate at build-time and instead focus
// on a lightweight runtime solution that is easier to use, as automatic as possible while still being flexible,
// and simple.
//
// Key design points:
//
// No build-time integration required, everything happens at runtime.
//
// During the request, anything can indicate that .. is required (by...)
//
// Requiring via a registry (as provided by the uiregistry package) is supported but not required for operation,
//     no direct dependency on it.
//
// The result is a single JS and a single CSS file which uniquely idenfies that set of files.  This can be disabled
//     but the idea is that you normally don't want to as this is a good way to do it - good browser performance and fully functional.
//
// The name (path) of that file should be deterministic - i.e. the same set of files should
//     always require in the same file name.  The name should also take into account the content of the
//     files (in case any of the dependencies cahnges it results in a new file name); but should note take into
//     account things which might be different across servers in a cluster, i.e. file timestamp.  The idea here is
//     that you can run this on a cluster and get predictable results without the nodes of the cluster requiring communication
//     or synchronization.  File timestamps can be used as a hint for cache invalidation but should have nothing to do with
//     the actual output.
//
// It should be possible for file names from old processes to still be served after a restart even if those contents
//     cannot be reproduced.  The situation is that if files are cached either in the browser, or with a proxying CDN or
//     anything else that might be in the request stream from client to server, you can end up with a situation where a deploy
//     can cause old cached pages to request an old CSS or JS file which no longer exists, thus breaking that page.  This is
//     should not be required for all deployments but it's vital we have an effective solution for this.
//
// Files should be minified to whatever extent is feasible.
//
// It must be fast, with appropriate caching wherever feasible.  The idea is you don't even notice it.
//
// While it's acknowledged that caching issues that apply to CSS and JS may also apply to other assets likes images,
//     it's much less of an issue we're going off the assumption that solutions related to other static files are more trouble
//     than their worth.  CSS and JS files are the nightmare that we solve here.
//
// Even with all that, it's also darned simple to use.  Just as important as handling the various case above is that you
//     don't want to leap out the window when trying to use it.
//
// TODO: explain why we ended up adding a big long thing that can be decoded into the FileSet (because of the edge case of a cluster
//     where node A generates the file but the request comes back to B due to round robin setup, B has no way to produce
//     the file without this; another edge case is it does this and the underlying files have been updated, in which case
//     we give the user of this lib the option to generate anyway or error; this looks to be the best possible thing that
//     handles the most cases without requiring communication between cluster nodes)  So to put it simply, two mechanisms:
//     hash is best but can fail if the specific node in the cluster didn't see it first; and token which can fail if it's
//     too long to fit in a valid URL (it may also not be desirable because of it's length and dev might turn it off
//     due to that).
//
// TODO: clean up the wording on the above after things work and we know what they are called - it's all theoretical
// but later we can say "Use YXZ to do ABC" instead of "It should be possible to do ABC".
//
// The trick with this dynamic file combination is to deal with the different cases that can arise between the time the
// page is served and when the request for the resource (the combined CSS/JS/etc) arrives - particularly in a cluster;
// where the version of your application may vary during deployment.  Some of these are edge cases but believe me
// they do happen and sometimes with, when unaccounted for, surprising results.  The objective of this package is to deal
// with these in a way that results in completely correct behavior in as many cases as possible and sensible behavior
// if correctness is not possible.  Some of the inspiration for this comes from Drupal's drupal_add_js() and their
// CSS/JS combiner, but their implementation is full of unaccounted for edge cases.
// Scenarios:
// 1. A page request arrives at node A, which generates a CSS/JS file.
//    This CSS/JS file is then requested back to the same server which can fulfill the request from in-memory cache.
//    The hash is used as the file name.
// 2. A page request arrives at node A but the server has restarted by the time the CSS/JS file request arrives at the same server.
//    This is handled by the disk cache - the proper response can be served from the file generated during the earlier run.
//    (Note that "disk cache" can be replaced with something that reads/writes to a CDN and thus allows that workflow if so
//    so desired.)
// 3. A page request arrives at node A but the request for the CSS/JS file arrives at node B, which has the same files necessary
//    to recreate the response, but does not have efficient way to compute that response only from the request path (filename, hash).
//    To solve this we need to provide additional information - the set of files in the original request.  This makes the path
//    much longer, but results in correct behavior in this scenario.  This additional piece of data we are calling a "token"
//    and it's an encoded list of the individual files that comprise the combined file.  Note that in cases where the path is very long it may be truncated
//    by the browser.  It may also just be undesirable to have that much additional data in the filename (referenced in the HTML page).
//    So there are tradeoffs but this is still a good solution for this case.
// 4. A page request arrives at node A but the request for the CSS/JS file arrives at node B which does not have the correct
//    version of the files (i.e. it gets the token and decodes it and tries to build the proper response but the hash
//    does not compute correctly - meaning whatever file contents were used when creating this request is not the file contents
//    we have).  This is the trickiest case, because it's impossible for the server to directly serve the right
//    response - it simply doesn't have the correct version of the files.  The most correct action would seem to be to have node B request
//    the correct data from node A.  The problem with this approach is complexity.
//    Orders of magnitude more code is required to properly implement this and making something that works
//    correctly for all or even a majority of deployment scenarios is not feasible.  Thus we support the following responses
//    in this case, each seemingly "less correct" but far simpler and more practical:
//    - Artificially block the request for a few seconds and then respond with a 307 back to the same URL, hoping that the
//      delay will allow time for further deployment and the
//      load balancer will pick a node capable of handling the request correctly (either this node after it has been updated
//      or another one that has already been updated).  The redirected URL also indicates how many
//      times the request has been re-requested so there is max number (let's say 5 for example).  When the max is exceeded
//      and the request has not be filled properly, fall back to one of the next responses.
//    - Return whatever version of the data you have and tell the browser not to cache it.  This results in an unpredictable
//      experience for the user, but this is also an edge case and so a few individual responses like this are often acceptable,
//      even if not ideal.
//    - Respond with an error.  In this case you'd rather have the page broken than be wrong/out of date.
//      This might make sense in cases where an intricate application must guarantee correctness.  However for most sites this
//      is probably overkill.
//    The default behavior is to combine the first and second approach - block and redirect back to itself, if the max retries is
//    exceeded then return the data you have, telling the browser not to cache it.  If your deployment process takes less time
//    than the total block+redirect cycle to run it's course, your responses for updated pages will be 100% correct, if briefly
//    delayed in loading.  If it takes longer
// 5. The edgy-est edge case is if the page request is done against the old version and the CSS/JS file request
//    arrives against the new version.  If browser caching is disabled for HTML pages (recommended,
//    see https://developers.google.com/web/fundamentals/performance/optimizing-content-efficiency/http-caching), then this is much
//    less likely to happen - the CSS/JS file would be out of date, not newer than when the page was generated.  This effectively
//    works like case 4 above except that it may never get the right answer because the version that is being asked for has now
//    been replaced and isn't coming back.  So in short, if your deployment process is fast, and your pages are
//    "Cache-control: no-store" then you have a low chance of this happening but it's not impossible.
//
// TODO: example usage (simple as possible, no dependencies, or additional with more complete caveman setup)
package uifiles

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gocaveman/caveman/webutil"
)

type UIResolver interface {
	Lookup(name string) (webutil.DataSource, error)
	ResolveDeps(name ...string) ([]string, error)
}

// type UIRequirer interface {
// 	// name should be in UI registry format if it starts with a slash
// 	// it means it's a local path, e.g. "js:/path/to/local.js" or "js:github.com/whatever/lib"
// 	UIRequire(name string) error
// }

func NewFileMangler(uiResolver UIResolver, localFs http.FileSystem, outputStore OutputStore) *FileMangler {

	ret := &FileMangler{
		URLPrefix:      "/fm-assets",
		UIResolver:     uiResolver,
		LocalFilesFs:   localFs,
		OutputStore:    outputStore,
		Minifier:       NewDefaultMinifier(),
		hashContentMap: make(map[string][]byte),
		prehashHashMap: make(map[string]string),
	}
	return ret
}

// FileMangler is a handler which assigns a FileSet to each request that runs through it.
// The FileSet can then be used to require CSS and JS files and get a URL that corresponds
// to that set for inclusion in your page.
type FileMangler struct {
	URLPrefix string // URL prefix to use, default from NewFileMangler is "/fm-assets"

	UIResolver   UIResolver      // for resolving libraries (usually from registry)
	LocalFilesFs http.FileSystem // for resolving local files not via registry
	// LocalFilesHandler http.Handler    // TODO: later, local files can also be obtained by internal request (need to decide what has higher priority); but some things to resolve like how to tell if a file is changed without doing a full request (HEAD?)

	Minifier Minifier // the minifier to use (default/nil will use github.com/tdewolff/minify)

	// TokenKey is used to encode the token, if empty then the use of tokens is disabled (i.e. not generated
	// in links and incoming ones are ignored).
	// Setting this to a specific value can be used to ensure tokens are the same across process restarts, recommended
	// for nodes in a cluster.
	// It does not have any specific length requirements but 32 bytes is recommended.  Do not use the same key for
	// sessions or other sensitive uses, generate something separate for this.
	TokenKey []byte

	// Gets called in situations where the wrong version of a file would be returned.
	// If the handler writes a response (implies cancelling the request context), then the wrong content will not
	// be returned.  Otherwise if no action is taken by this handler the wrong content will be returned.
	WrongContentHandler http.Handler

	hashContentMap map[string][]byte // cache of hash -> content
	prehashHashMap map[string]string // cache of prehash -> hash
	mu             sync.RWMutex      // lock for maps

	// storage for output files, usually local directory impl with FileSystemOutputStore but could also
	// be backed with remote storage (CDN for example)
	OutputStore OutputStore

	OutputStoreGCDelay time.Duration // how long do we way between calls to OutputStore GC, default if 0 is 10 min

	outputStoreGCInProgress bool         // are we doing OutputStore GC
	outputStoreGCLastTime   time.Time    // when did we last OutputStore GC
	outputStoreGCmu         sync.RWMutex // lock specifically for OutputStore garbage collection

}

func (fm *FileMangler) resolveFileDataSources(files FileEntryList) ([]webutil.DataSource, error) {
	var ret []webutil.DataSource
	for _, f := range files {
		if fm.UIResolver == nil {
			return nil, fmt.Errorf("no resolver, can't look up DataSource")
		}
		ds, err := fm.UIResolver.Lookup(f)
		if err != nil {
			// Note that resolveFileDataSources should never be called on a resource that doesn't exist,
			// since the definition of a "resolve file" vs a "local file" is if the resolved knows about
			// it (i.e. we got a valid Lookup() on it before) - see FileSet.resolveDeps;
			// Thus we don't need to check for the "not found" condition here and do something special with it
			return nil, err
		}
		ret = append(ret, ds)
	}
	return ret, nil
}

func (fm *FileMangler) localFileDataSources(files FileEntryList) ([]webutil.DataSource, error) {

	var missing []string

	var ret []webutil.DataSource
	for _, f := range files {

		fparts := strings.SplitN(f, ":", 2)
		fname := fparts[len(fparts)-1]

		ffile, err := fm.LocalFilesFs.Open(fname)
		if os.IsNotExist(err) || err == webutil.ErrNotFound {
			missing = append(missing, fname)
		} else {
			ffile.Close()

			ret = append(ret, webutil.NewHTTPFSDataSource(fm.LocalFilesFs, fname))
		}

	}

	// if any missing files we still return our result but with an error that indicates the condition
	var reterr error
	if len(missing) > 0 {
		reterr = NewMissingRequirementsError(missing)
	}

	return ret, reterr
}

func (fm *FileMangler) runGCIfNeeded() {

	// figure out the delay
	delay := fm.OutputStoreGCDelay
	if delay <= 0 {
		delay = time.Minute * 10
	}

	// see if time exceeded and we're not already running
	needsGC := false
	fm.outputStoreGCmu.RLock()
	if (!fm.outputStoreGCInProgress) && time.Now().After(fm.outputStoreGCLastTime.Add(delay)) {
		needsGC = true
	}
	fm.outputStoreGCmu.RUnlock()

	// log.Printf("needsGC=%v fm.outputStoreGCLastTime=%v", needsGC, fm.outputStoreGCLastTime)

	// if GC needed acquire write lock for that and proceed
	if needsGC {
		// log.Printf("needsGC")
		fm.outputStoreGCmu.Lock()
		// double check to make sure it's still needed
		if (!fm.outputStoreGCInProgress) && time.Now().After(fm.outputStoreGCLastTime.Add(delay)) {
			// update timestamp and kick it off in a separate goroutine
			fm.outputStoreGCLastTime = time.Now()
			fm.outputStoreGCInProgress = true
			go func() {
				// log.Printf("performing GCFiles")
				err := fm.OutputStore.GCFiles()
				if err != nil {
					// nothing much we can do except log it
					log.Printf("Error during FileMangler.OutputStore.GCFiles: %v", err)
				}
				fm.outputStoreGCmu.Lock()
				defer fm.outputStoreGCmu.Unlock()
				fm.outputStoreGCInProgress = false
			}()
		}
		fm.outputStoreGCmu.Unlock()
	}

}

func (fm *FileMangler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// check for path prefix
	p := path.Clean("/" + r.URL.Path)
	if !strings.HasPrefix(p, fm.URLPrefix+"/") {
		return
	}

	// run GC if it's time
	fm.runGCIfNeeded()

	lp := strings.TrimPrefix(p, fm.URLPrefix+"/")

	// serve combined file
	// note probably an exact regexp match on name (exact hash length, single dot, then file ext) would work well here,
	// although I don't see a reason this logic wouldn't work either...
	if !strings.Contains(lp, "/") {

		lpparts := strings.Split(lp, ".")
		if len(lpparts) != 2 {
			http.NotFound(w, r)
			return
		}
		hashString := lpparts[0]
		typeName := lpparts[1]

		fm.mu.RLock()
		content, ok := fm.hashContentMap[hashString]
		fm.mu.RUnlock()

		fname := hashString + "." + typeName

		// not in memory, let's try OutputStore
		if !ok {
			if fm.OutputStore != nil {
				b, err := fm.OutputStore.ReadFile(fname)
				if err != nil {
					if err != webutil.ErrNotFound {
						log.Printf("FileMangler: Error reading from OutputStore(fname=%q): %v", fname, err)
					}
				} else {
					content = b
				}
			}
		}

		// if we didn't get it by now, let's try recreating it from the token
		token := r.URL.Query().Get("t")
		if content == nil && len(fm.TokenKey) > 0 && token != "" {

			var fileEntryList FileEntryList
			err := DecodeToken(&fileEntryList, fm.TokenKey, token)
			if err != nil {
				log.Printf("Error decoding token (%q): %v", token, err)
			} else {

				fileSet := &FileSet{fileMangler: fm, files: fileEntryList}

				setp, err := fileSet.BuildSetPathNoPrefix(typeName)
				if err != nil {
					// specifically check for IsMissingResourceError and use that to trigger the WrongContentHandler
					if IsMissingRequirementsError(err) {
						if fm.WrongContentHandler != nil {
							w, r = webutil.ServeHTTPChain(fm.WrongContentHandler, w, r)
							if r.Context().Err() != nil { // is request handled
								return
							}
						}
						// if WrongContentHandler didn't exist or didn't handle above, fall through
					} else {
						// any other error we bail on
						webutil.HTTPError(w, r, err, "internal error", 500)
						return
					}
				}

				// see if what was generated has the same hash, if not then something changed since the URL we're
				// serving was generated
				if !strings.HasPrefix(setp, hashString) {

					if fm.WrongContentHandler != nil {
						w, r = webutil.ServeHTTPChain(fm.WrongContentHandler, w, r)
						if r.Context().Err() != nil { // is request handled
							return
						}
					}

					// not handled, in this case we just proceed to serve what was just created
					// and log a warning
					newHashString := strings.SplitN(setp, ".", 2)[0]
					abbrevToken := token
					if len(abbrevToken) > 40 {
						abbrevToken = abbrevToken[:40] + "..."
					}
					log.Printf("Warning - FileMangler: expected hash %q but after generating from token %q got %q; falling through to serve what we have", hashString, abbrevToken, newHashString)
					hashString = newHashString
					// disable cache in this case - we don't want browsers hanging onto the wrong stuff
					w.Header().Set("cache-control", "no-store")

				}

				// at this point, whatever happend above should have given us a hashString which is the content to serve
				fm.mu.RLock()
				content, _ = fm.hashContentMap[hashString]
				fm.mu.RUnlock()

			}

		}

		// we have it in memory cache, so just serve it
		if content != nil {
			header := w.Header()
			// if no cache-control set, our behavior is to tell the browser to cache for a long time
			if header.Get("cache-control") == "" {
				header.Set("cache-control", "max-age=604800") // 1 week
			}
			if header.Get("content-type") == "" {
				ct := mime.TypeByExtension("." + typeName)
				if ct != "" {
					w.Header().Set("content-type", ct)
				}
			}

			// TODO: if we can determine an appropriate date, we may be better off using
			// http.ServeContent here, in order to avoid sending back the guts of the file
			// if not needed (i.e. the case of if-not-modifed header and 304 response)
			w.Write(content)

			// kick off a touch in the background (trying to keep in-memory cache hits as fast as possible)
			if fm.OutputStore != nil {
				go func() {
					fname := hashString + "." + typeName
					err := fm.OutputStore.TouchFile(fname)
					if err != nil {
						log.Printf("OutputStore.TouchFile(%q) error: %v", fname, err)
					}
				}()
			}

			return
		}

		return

	}

	// now check for individual file

	// Make this super obvious and direct, no translation at all:
	// js:github.com/gocaveman-libs/jquery -> /fm-assets/js:github.com/gocaveman-libs/jquery
	// js:/path/to/local.js -> /fm-assets/js:/path/to/local.js

	resourceName := lp
	ver := r.URL.Query().Get("ver")

	ds, err := fm.UIResolver.Lookup(resourceName)
	if err == webutil.ErrNotFound {
		resourceNameParts := strings.SplitN(resourceName, ":", 2)
		if len(resourceNameParts) > 1 {
			fname := resourceNameParts[1]
			localF, err := fm.LocalFilesFs.Open(fname)
			if err == nil {
				localF.Close()
				// FIXME: this is kinda messy - we open the file twice, should be a good way to
				// detect if file doesn't exist but not have to reopen it - maybe there's a variation
				// on NewHTTPFSDataSource that can accept an open file - maybe it's an "open once data source"
				ds = webutil.NewHTTPFSDataSource(fm.LocalFilesFs, fname)
			}
		}
	}

	if err == webutil.ErrNotFound || os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		webutil.HTTPError(w, r, err, "internal error finding file", 500)
		return
	}

	rsc, err := ds.OpenData()
	if err != nil {
		webutil.HTTPError(w, r, err, "internal error opening data source", 500)
		return
	}
	defer rsc.Close()

	// read the data in
	b, err := ioutil.ReadAll(rsc)
	if err != nil {
		webutil.HTTPError(w, r, err, "internal error reading data source", 500)
		return
	}

	// hash it
	hash := NewHash()
	hash.Write(b)
	hashString := hash.ResultString()

	// compare hash and use it to set header
	if hashString == ver {
		// keep it for a week
		w.Header().Set("cache-control", "max-age=604800")
	} else {
		// don't store it at all - it's the wrong version
		w.Header().Set("cache-control", "no-store")
	}

	if w.Header().Get("content-type") == "" {
		ct := mime.TypeByExtension("." + strings.SplitN(resourceName, ".", 2)[0])
		if ct != "" {
			w.Header().Set("content-type", ct)
		}
	}

	// try to get a valid mod time so we can do http.ServeContent
	var modTime time.Time
	st, _ := ds.Stat()
	if st != nil {
		modTime = st.ModTime()
	}

	if !modTime.IsZero() {
		http.ServeContent(w, r, resourceName, modTime, bytes.NewReader(b))
	} else {
		w.Write(b)
	}

	return

}

func (fm *FileMangler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wnext http.ResponseWriter, rnext *http.Request) {

	fm.ServeHTTP(w, r)

	// attach to context
	ctx := r.Context()
	fileSet := &FileSet{
		fileMangler: fm,
	}
	ctx = CtxWithFileSet(ctx, fileSet)

	return w, r.WithContext(ctx)

}

func CtxFileSet(ctx context.Context) *FileSet {
	ret, _ := ctx.Value("uifiles.FileSet").(*FileSet)
	return ret
}

func CtxWithFileSet(ctx context.Context, fs *FileSet) context.Context {
	return context.WithValue(ctx, "uifiles.FileSet", fs)
}

// Require provides a one-line way to add a UI dependency.  Any error is returned.
func Require(r *http.Request, name string) error {
	fileSet := CtxFileSet(r.Context())
	if fileSet == nil {
		return fmt.Errorf("Request context does not contain a uifiles.FileSet, FileManager is probably missing from your handler chain")
	}
	return fileSet.Require(name)
}

// MustRequire provides a one-line way to add a UI dependency.  If it fails
// webutil.HTTPError will be used to log the error and output the appropriate error to the ResponseWriter
func MustRequire(w http.ResponseWriter, r *http.Request, name string) {
	err := Require(r, name)
	if err != nil {
		webutil.HTTPError(w, r, err, fmt.Sprintf("MustRequire error for %q", name), 500)
		return
	}
}

type FileSet struct {
	fileMangler *FileMangler
	// files       []string
	files FileEntryList
}

func (fs *FileSet) Require(name string) error {

	// check for dup
	for _, n := range fs.files {
		if n == name {
			return nil
		}
	}

	// first check name against UIResolver
	if fs.fileMangler.UIResolver != nil {
		_, err := fs.fileMangler.UIResolver.Lookup(name)
		if err != nil {
			if err != webutil.ErrNotFound {
				return err // return error other than not found
			}
		} else {
			// the UIResolver found it, we're good
			fs.files = append(fs.files, name)
			return nil
		}
	}

	// now check against local file system
	nameParts := strings.SplitN(name, ":", 2)
	if len(nameParts) < 2 {
		return fmt.Errorf("name %q does not have type", name)
	}
	f, err := fs.fileMangler.LocalFilesFs.Open(nameParts[1])
	if err != nil {
		return err
	}
	f.Close()

	fs.files = append(fs.files, name)

	return nil
}

// resolveDeps resolves the dependencies for the current set of files
// and returns the set broken up into files retrieved through the resolver
// and files that need to be retrieved as local files.  If an error occurs
// during resolution
func (fs *FileSet) resolveDeps() (resolveFiles FileEntryList, localFiles FileEntryList, reterr error) {

	uiResolver := fs.fileMangler.UIResolver

	for _, f := range fs.files {

		if uiResolver != nil {

			// see if the UIResolver knows about this file
			_, err := uiResolver.Lookup(f)
			if err == nil {
				resolveFiles = append(resolveFiles, f)
				continue
			} else if err != nil && err != webutil.ErrNotFound {
				return nil, nil, fmt.Errorf("error while resolving %q: %v", f, err)
			}

		}

		localFiles = append(localFiles, f)

	}

	// reassemble fs.files with the resolve stuff first an the locals later
	fs.files = make(FileEntryList, 0, len(resolveFiles)+len(localFiles))
	fs.files = append(fs.files, resolveFiles...)
	fs.files = append(fs.files, localFiles...)

	resolveFiles, reterr = uiResolver.ResolveDeps(resolveFiles...)

	return

}

func (fs *FileSet) FilePaths(filterType string) ([]string, error) {

	files, err := fs.FilePathsNoPrefix(filterType)
	if err != nil {
		return nil, err
	}

	var ret []string

	for _, f := range files {
		ret = append(ret, fs.fileMangler.URLPrefix+"/"+f)
	}

	return ret, nil

}

// BuildFilePaths is the unminified and uncombined set of files, but in the correct sequence and ready to be output.
// It's intended for debugging JS and/or CSS.
// TODO: we really should track back to the configuration how this
// gets used instead of BuildSetPath when needed.  The ideal way is that
// in config you can (and it can be overridden based on deployment context)
// set a flag which the template would pick up to use this call instead of
// BuildSetPath.  But BuildSetPath would be the default if unspecified.
func (fs *FileSet) FilePathsNoPrefix(filterType string) ([]string, error) {

	// use uiResolver to add depedencies and sort and to separate into resolveFiles and localFiles
	resolveFiles, localFiles, err := fs.resolveDeps()
	if err != nil {
		return nil, err
	}

	// filter by type
	resolveFiles = fileEntriesWithType(resolveFiles, filterType)
	localFiles = fileEntriesWithType(localFiles, filterType)

	var allDSs []webutil.DataSource
	resolveDSs, err := fs.fileMangler.resolveFileDataSources(resolveFiles)
	if err != nil {
		return nil, err
	}
	allDSs = append(allDSs, resolveDSs...)
	localDSs, err := fs.fileMangler.localFileDataSources(localFiles)
	if err != nil {
		return nil, err
	}
	allDSs = append(allDSs, localDSs...)

	// this should be impossible but put in a check to ensure sanity
	if len(allDSs) != len(resolveFiles)+len(localFiles) {
		return nil, fmt.Errorf("unexpected state len(allDSs)[%d] != len(resolveFiles)[%d] + len(localFiles)[%d]", len(allDSs), len(resolveFiles), len(localFiles))
	}

	// for now, since this call is intended for development, we're going to rehash
	// these files each time, ideally we would use a prehash to cache it
	var ret []string

	for i := 0; i < len(resolveDSs); i++ {
		ds := resolveDSs[i]
		name := resolveFiles[i]
		rsc, err := resolveDSs[i].OpenData()
		if err != nil {
			return nil, fmt.Errorf("FilePathsNoPrefix error opening resolve file (ds=%+v): %v", ds, err)
		}
		defer rsc.Close()
		hash := NewHash()
		_, err = io.Copy(hash, rsc)
		if err != nil {
			return nil, fmt.Errorf("FilePathsNoPrefix error reading resolve file (ds=%+v): %v", ds, err)
		}
		hashString := hash.ResultString()
		ret = append(ret, name+"?ver="+hashString)
	}

	for i := 0; i < len(localDSs); i++ {
		ds := localDSs[i]
		name := localFiles[i]
		rsc, err := localDSs[i].OpenData()
		if err != nil {
			return nil, fmt.Errorf("FilePathsNoPrefix error opening local file (ds=%+v): %v", ds, err)
		}
		defer rsc.Close()
		hash := NewHash()
		_, err = io.Copy(hash, rsc)
		if err != nil {
			return nil, fmt.Errorf("FilePathsNoPrefix error reading local file (ds=%+v): %v", ds, err)
		}
		hashString := hash.ResultString()
		ret = append(ret, name+"?ver="+hashString)
	}

	return ret, nil
}

func (fs *FileSet) BuildSetPath(filterType string) (string, error) {
	ret, err := fs.BuildSetPathNoPrefix(filterType)
	return fs.fileMangler.URLPrefix + "/" + ret, err
}

// BuildSetPath performs file combination and whatever else to build the specified set and returns the path name (intended for output in the HTML page).
// You can call BuildSetPath as many times as you want for a given set but once it is called you must not call UIRequire() again for this FileSet (for this HTTP request).
// If everything is cached, this operation could return very quickly.
// The setName is usually either "js" or "css".
func (fs *FileSet) BuildSetPathNoPrefix(filterType string) (ret string, reterr error) {

	// use uiResolver to add depedencies and sort and to separate into resolveFiles and localFiles
	resolveFiles, localFiles, err := fs.resolveDeps()
	if err != nil {
		return "", err
	}

	// filter by type
	resolveFiles = fileEntriesWithType(resolveFiles, filterType)
	localFiles = fileEntriesWithType(localFiles, filterType)
	bothFiles := make(FileEntryList, 0, len(resolveFiles)+len(localFiles))
	bothFiles = append(bothFiles, resolveFiles...)
	bothFiles = append(bothFiles, localFiles...)

	// calculate the token suffix part if needed - we're going to need it later
	tokenSuffix := ""
	if len(fs.fileMangler.TokenKey) > 0 {
		t, err := EncodeToken(bothFiles, fs.fileMangler.TokenKey)
		if err != nil {
			return "", err
		}
		tokenSuffix = "?t=" + t
	}

	// get all of the DataSources for resolveFiles and all of the http.Files for localFiles
	// FIXME: from resolveFileDataSources and localFileDataSources we should detect the case
	// of a missing file and build all the stuff we can and return the answer and that error,
	// in ServeHTTP we need to detect that case, whereas in the template the error will show on the page (as it should)
	var allDSs []webutil.DataSource
	resolveDSs, err := fs.fileMangler.resolveFileDataSources(resolveFiles)
	if err != nil {
		return "", err
	}
	allDSs = append(allDSs, resolveDSs...)
	localDSs, err := fs.fileMangler.localFileDataSources(localFiles)
	if err != nil {

		// missing requirements we continue on and pass the error back to the caller
		if IsMissingRequirementsError(err) {
			reterr = err
		} else {
			// anything else we bail
			return "", err
		}

	}
	allDSs = append(allDSs, localDSs...)

	// use the file names and timestamps to generate a "prehash"
	prehash := NewHash()
	prehash.FileEntryList(resolveFiles)
	prehash.FileEntryList(localFiles)
	// TODO: we could/should add an option to skip this step - this would mean that
	// in production you can avoid hitting the disk to check the timestamps every time a request is made
	// (although the rest of the code path needs to be checked to ensure we're not incidentally hitting
	// it somewhere else - but still should be doable in theory)
	prehash.DataSourceModTimes(resolveDSs)
	prehash.DataSourceModTimes(localDSs)
	prehashString := prehash.ResultString()

	// look up prehash in a local in memory map of prehash->hash, if found, just return hash (filename)
	fs.fileMangler.mu.RLock()
	hashString := fs.fileMangler.prehashHashMap[prehashString]
	fs.fileMangler.mu.RUnlock()
	if hashString != "" {
		return hashString + "." + filterType + tokenSuffix, nil
	}

	// no prehash entry found, we need to regenerate everything

	var w bytes.Buffer
	for _, ds := range allDSs {

		rsc, err := ds.OpenData()
		if err != nil {
			return "", err
		}
		defer rsc.Close()

		if fs.fileMangler.Minifier != nil {
			err = fs.fileMangler.Minifier.Minify(filterType, &w, rsc)
			if err != nil {
				return "", fmt.Errorf("Error during minification (DataSource=%+v): %v", ds, err)
			}
		} else {
			_, err = io.Copy(&w, rsc)
			if err != nil {
				return "", err
			}
		}

		w.WriteByte('\n') // blank line between each file

	}

	// - calc content hash
	wb := w.Bytes()
	hash := NewHash()
	hash.Write(wb)
	hashString = hash.ResultString()

	// at this point we need to write lock everyone else out
	fs.fileMangler.mu.Lock()
	defer fs.fileMangler.mu.Unlock()

	// store content keyed by hash
	fs.fileMangler.hashContentMap[hashString] = wb

	// write to output store file if enabled
	if fs.fileMangler.OutputStore != nil {
		err := fs.fileMangler.OutputStore.WriteFile(hashString+"."+filterType, wb)
		if err != nil {
			return "", err
		}
	}

	// store prehash->hash
	fs.fileMangler.prehashHashMap[prehashString] = hashString

	// we're good now
	ret = hashString + "." + filterType + tokenSuffix

	// return, including any reterr that may have been set above
	return ret, reterr

}

// NewMissingRequirementsError does what you think it does.
func NewMissingRequirementsError(what []string) error {
	return MissingRequirementsError{Requirements: what}
}

// MissingRequirementsError is used to indicate that one or more requirements aren't there.
// We need this separated out because it's an edge case that can happen during deployment
// or with cached pages and we want predictable behavior.
type MissingRequirementsError struct {
	Requirements []string
}

func (e MissingRequirementsError) Error() string {
	return fmt.Sprintf("Missing requirements: %+v", e.Requirements)
}

func IsMissingRequirementsError(err error) bool {
	_, ok := err.(MissingRequirementsError)
	return ok
}

type FileEntryList []string

func fileEntriesWithType(l FileEntryList, t string) FileEntryList {
	ret := make(FileEntryList, 0, len(l))
	for _, e := range l {
		if strings.HasPrefix(e, t+":") {
			ret = append(ret, e)
		}
	}
	return ret
}
