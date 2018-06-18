// Provides YAML parsing for use with i18n package.
//
// YAML files are expected to be a simple list of name: value pairs.
// The group and locale either provided during the appropriate Load call
// or infered from the file name for functions which support that.
// File name convention is group.locale.yaml, e.g. "default.en-gb.yaml"
// contains text for the "default" group and "en-gb" locale. Multiple
// files can be provided for the same group by using the form:
// "group-additional.local.yaml".  Example: "default-set1.en-gb.yaml",
// "default-set2.en-gb.yaml" - with the "-set1" part being ignored
// and only used to ensure the files sort properly - later ones taking
// higher priority.
//
package i18nyaml

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/gocaveman/caveman/i18n"
	"github.com/gocaveman/caveman/webutil"
	yaml "gopkg.in/yaml.v2"
)

// this should create a namedsequence with the paths as names
func LoadDir(fs http.FileSystem, dirpath string) (webutil.NamedSequence, error) {

	dirf, err := fs.Open(dirpath)
	if err != nil {
		return nil, err
	}
	defer dirf.Close()

	var ret webutil.NamedSequence

	fis, err := dirf.Readdir(-1)
	if err != nil {
		return nil, err
	}

	// sort by name reverse - later files get higher priority (lower sequence number)
	sort.Slice(fis, func(i, j int) bool {
		return fis[i].Name() >= fis[j].Name()
	})

	// explicit sequence number to preserve sequence from file system
	seqn := float64(0)

	for _, fi := range fis {
		seqn += 0.00001

		baseName := path.Base(fi.Name())

		g, l, ok := FileNameParse(baseName)
		if !ok {
			continue
		}

		f, err := fs.Open(path.Join(dirpath, baseName))
		if err != nil {
			return ret, err
		}
		defer f.Close()

		tr, err := Load(f, g, l)
		if err != nil {
			return ret, err
		}

		ret = append(ret, webutil.NamedSequenceItem{Sequence: 50 + seqn, Name: fi.Name(), Value: tr})
	}

	return ret, nil
}

func FileNameParse(fn string) (group, locale string, ok bool) {
	fn = path.Base(fn)
	parts := strings.Split(fn, ".")
	if len(parts) != 3 {
		return "", "", false
	}

	log.Printf("FIXME: need to support group-additional")
	group = parts[0]
	locale = parts[1]
	ok = parts[2] == "yaml"
	return
}

// Load reads a file containing YAML and returns a i18n.MapTranslator with the result.
func LoadFile(fpath, g, l string) (*i18n.MapTranslator, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Load(f, g, l)
}

// Load reads a Reader containing YAML and returns a i18n.MapTranslator with the result.
func Load(r io.Reader, g, l string) (*i18n.MapTranslator, error) {

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	ms := make(map[string]string)

	err = yaml.Unmarshal(b, &ms)
	if err != nil {
		return nil, err
	}

	ret := i18n.NewMapTranslator()
	for k, v := range ms {
		ret.SetEntry(g, k, l, v)
	}

	return ret, nil
}

// Hm, should we have tooling here to go back and forth between db?  This would allow us to
// use sqlite for local instances for the editor, and other db for other cases but then
// just rip through using Gorm and write out to file(s).
