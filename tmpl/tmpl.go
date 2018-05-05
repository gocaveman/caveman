package tmpl

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/gocaveman/caveman/webutil"
)

const (
	ViewsCategory    = "views"    // default category for view templates
	IncludesCategory = "includes" // default category for include templates
)

const (

	// GoTemplateHtmlMimeType is the mime type for HTML templates
	// If there is a more appropriate one to use, by all means someone let me know, I couldn't find it.
	GoTemplateHtmlMimeType = "text/gotemplate+html"

	// MarkdownMimeType is the mime type for markdown files
	MarkdownMimeType = "text/markdown"
)

// DefaultFileExtMimeTypes is a map of mime types we know the
// template file extension mappings for by default.
var DefaultFileExtMimeTypes = map[string]string{
	".html":   GoTemplateHtmlMimeType,
	".gohtml": GoTemplateHtmlMimeType,
	".md":     MarkdownMimeType,
}

// ErrNotFound indicates the item isn't there.
var ErrNotFound = webutil.ErrNotFound

var ErrAlreadyExists = errors.New("already exists")

// ErrWriteNotSupported indicates the Store does not support writing,
var ErrWriteNotSupported = errors.New("write not supported")

// Store is a storage for templates.  Templates are organized into categories,
// the common cases are "views", intended for direct rendering, and
// "includes", intended to be included.  Templates can be stored in any format
// and it is the Renderer's (different package) responsibilty to adapt it to
// an actual html/template.  A Store can be implemented on top of a filesystem,
// a database table, in memory, or anything else.  It can also be read-only,
// in which case its write operations will return ErrWriteNotSupported.
type Store interface {

	// Write a new template, will return ErrAlreadyExists for existing template.
	CreateTemplate(category, fileName string, body []byte, mimeType string, meta map[string]interface{}) error

	// Read a template, returning a complete Tmpl with Body and Meta.
	// ErrNotFound will be returned if it doesn't exist.
	ReadTemplate(category, fileName string) (body []byte, mimeType string, meta map[string]interface{}, err error)

	// Write a template.  It must exist first or or ErrNotFound will be returned.
	// The Meta will be merged into the existing template as part of the update.
	UpdateTemplate(category, fileName string, body []byte, mimeType string, meta map[string]interface{}) error

	// Remove a template.  If it did not exist, ErrNotFound will be returned.
	DeleteTemplate(category, fileName string) error

	// Categories returns the list of available categories.
	Categories() ([]string, error)

	// FindByPrefix will return a slice of file names for the specified category
	// that begin with a prefix, up to an indicated limit.  Limit <= 0 means all.
	// Prefix "/" will list all.  All Stores must support "/directory/" as a prefix
	// to read below a specific directory, behavior for prefixes which do not end
	// in a slash is implementation-specific.
	FindByPrefix(category, fileNamePrefix string, limit int) ([]string, error)
}

// TemplateReader is a subset of Store that only has ReadTemplate.
type TemplateReader interface {
	ReadTemplate(category, fileName string) (body []byte, mimeType string, meta map[string]interface{}, err error)
}

// // Tmpl is a single template
// type Tmpl struct {
// 	Category     string
// 	FileName     string
// 	Body         []byte
// 	BodyMimeType string // mime type of input, use "text/gotemplate+html" for Go's html/template, "text/markdown" for markdown
// 	Meta         StringDataMap
// }

// ParseYAMLHeadTemplate reads data and splits off the meta and parses into
// YAMLStringDataMap and returns that and the body contents.
func ParseYAMLHeadTemplate(in io.Reader) (meta *YAMLStringDataMap, body []byte, err error) {

	var tmplPrefix bytes.Buffer
	var yamlPart bytes.Buffer

	br := bufio.NewReader(in)

	line, err := br.ReadString('\n')
	if err != nil {
		return nil, nil, err
	}
	// see if the first line has the YAML separator
	if strings.TrimRight(line, "\t\r\n ") == "---" {

		yamlPart.WriteString(line)

		// replace separator with go template open comment
		tmplPrefix.WriteString("{{/*\n")

		for {
			line, err := br.ReadString('\n')
			if err == io.EOF {
				return nil, nil, fmt.Errorf("file starts a block with '---' but does not close it")
			} else if err != nil {
				return nil, nil, err
			}

			// check for end of block
			if strings.TrimRight(line, "\t\r\n ") == "---" {
				tmplPrefix.WriteString("\n*/}}") // write the newline before in order to ensure templates can have no whitespace in them at the beginning if needed (doctype, e.g.)
				break
			} else {
				// not end of block, just add a blank line (keeping the line count the same for easier debugging)
				tmplPrefix.WriteString("\n")
				yamlPart.WriteString(line) // record yamlPart for parsing
			}
		}

	} else {
		// no YAML separator, take the line we just read and put in the buffer and continue
		tmplPrefix.WriteString(line)
	}

	// read everything else
	rest, err := ioutil.ReadAll(br)
	if err != nil {
		return nil, nil, err
	}
	// put it after the prefix
	tmplPrefix.Write(rest)
	// that's our return body
	body = tmplPrefix.Bytes()

	// if any YAML, parse it
	if yamlPart.Len() > 0 {
		retMap, err := ReadYAMLStringDataMap(bytes.NewReader(yamlPart.Bytes()))
		if err != nil {
			return nil, nil, err
		}
		meta = retMap
	}

	return
}
