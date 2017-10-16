package renderer

// some specific things to break out:
// - path <-> filename mapping, needs to be bidirectional and pluggable
//   filenames can be multiple options
// - file contents -> go template conversion (need a default one plus funky stuff like markdown);
//   also should support redirects, e.g. /index -> / (but also look at this in relation to the
//   /section/ vs /section issue, which may not be ble to be handled here but might require a separate 404
//   handler - sort out the difference...)
// - what do we do about metadata... that one is odd, i'm tempted to completely leave it out of this step...
//   it could be that the pages module thing runs before and picks the metadata from whatever it's doing
//   and attaches it to the context and that's it, renderer has nothing to do with it.
// - require functionality - but this should be pluggable, rather than having a UIRequirer right on Renderer,
//   we could do some sort of generic post processor (after parse and before exec) that is enabled by default
// - shoudl we have a way to customize the template that gets created (i.e. to define other templates
//   or... )?  possibly, but without understanding
//   the use this is likely overkill for a first version - could easily add later.

func New() /* ... */ {

}

type FileNamer interface {
	FileNames(path string) []string // hm, do we support redirects here...?
}

type Pather interface {
	Path(filename string) string
}
