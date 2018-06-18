// Provides a Go interface, REST endpoints and a web UI for editing translations.
package i18neditor

//
// figure out what methods are needed for this concept of a set of files that correlates to translations
// and read/write for it
//
// maybe even something you can embed that handles the file stuff and you only have to implement the
// read/write for your specific format - could be very useful
//
// should also be able to serve as a Translator
//
// webserver/handler
//
// think about the workflow a bit: "what's left to translate" and similar situations
//
// hm - integration with google translate api would be really awesome here... at least make sure this is feasible to add
//

type Editor interface {
	// FilesForLocale() ([]string, error)
	Files() ([]FileInfo, error)
	FileInfoFor(path string) (FileInfo, error)
	FileContentsFor(path string) (*FileContents, error)
	WriteFileRecords(path string, fileRecords []FileRecord) error
}

type FileInfo struct {
	Path   string
	Group  string
	Locale string
}

type FileContents struct {
	FileRecords []FileRecord
}

type FileRecord struct {
	ID      string // surrogate key in case it's neded, can describe location in XLIFF file, etc.
	Key     string
	Value   string
	Data    string // ? (additional data as JSON?)
	Comment string
}

// TODO: add REST server that exposes an Editor
