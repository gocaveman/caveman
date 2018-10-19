package gen

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/jinzhu/inflection"
	"github.com/spf13/pflag"
)

// ideas of what we want to accomplish:

// cavegen ctrl-rest-crud src/mypjt/ctrl/ctrl-customer.go - probably need a -store CustomerStore option

// cavegen store-struct src/mypjt/store/store.go - the Store struct itself and setup

// cavegen store-crud src/mypjt/store/store-customer.go - CRUD calls for a customer, as methods on Store

// cavegen store-struct-crud src/mypjt/store/customer.go - CustomerStore and crud calls

// cavegen model-sample-customer src/mypjt/store/model-customer.go - an example of a customer

// cavegen ctrl-sample-pages src/mypjt/ctrl/ctrl-customer-pages.go - sample of page data controller

// cavegen asset-package src/mypjt/views/assets.go - make a dir that, with go:generate, is packaged into a fs avail at runtime

// cavegen embed-tmpl src/mypjt/somefiles/embed.go - use vfsgen to package a directory up into http.FileSystem and tmpl.Store and register it

type Settings struct {
	WorkDir string // directory that all of the paths are relative to
	GOPATH  string // GOPATH as extracted from env
}

// RelativeToGOPATH returns "src/whatever" from "./whatever" given
// that WorkDir is GOPATH+"/src".  Basically, it takes a file name
// as provided by the user and gives you something relative to GOPATH.
// Returns error if p is not in GOPATH.  You should use this to convert
// user-provided file names into something you can reliably work with.
// Absolute paths are made relative and checked to ensure they are under
// GOPATH or will return error.
// If the path starts with a "./" or "../" it is interpreted as relative to the
// current working directory.
// Otherwise it is assumed to already be relative to GOPATH.
func (s *Settings) RelativeToGOPATH(p string) (retv string, rete error) {

	// origp := p
	// defer func() {
	// 	log.Printf("RelativeToGOPATH %q is returning (%q, %v)", origp, retv, rete)
	// }()

	if !filepath.IsAbs(s.GOPATH) {
		return "", fmt.Errorf("GOPATH %q is not an absolute path, cannont continue", s.GOPATH)
	}

	// it should be interpreted as relative, use WorkDir to make an absolute path.
	if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") {
		p = filepath.Join(s.WorkDir, p)
	}

	// if it's still not absolute, use GOPATH to make it so
	if !filepath.IsAbs(p) {
		p = filepath.Join(s.GOPATH, p)
	}

	ret, err := filepath.Rel(s.GOPATH, p)
	if err != nil {
		return "", err
	}

	// FIXME: this is a bit of a hack, but for now we use it to determine if
	// the resulting path is "outside of GOPATH"
	if strings.HasPrefix(ret, "..") {
		return "", fmt.Errorf("%q is not under GOPATH %q", p, s.GOPATH)
	}

	return ret, nil
}

// // FormatGoCodeFile is like FormatGoCode but replaces a file in-place rather than operating on the bytes.
// func (s *Settings) FormatGoCodeFile(filename string) error {
// 	panic("not implemented yet")
// }

// FormatGoCode will try to use the goimports command to format the specified Go code.
// If not available it will try gofmt.  You provide the logical file name and
// the source code itself.  The GOPATH from Settings is used.
func (s *Settings) FormatGoCode(filename string, src []byte) ([]byte, error) {

	var cmdPath string
	cmdPath, err := exec.LookPath("goimports")
	if err != nil {
		cmdPath, err = exec.LookPath("gofmt")
		if err != nil {
			return nil, err
		}
	}

	// FIXME: ideally we would be using -srcdir when we use goimports
	cmd := exec.Command(cmdPath)
	cmd.Dir = s.WorkDir
	// copy os environment but use our own GOPATH
	env := os.Environ()
	didGoPath := false
	for i := range env {
		if strings.HasPrefix(env[i], "GOPATH=") {
			env[i] = "GOPATH=" + s.GOPATH
			didGoPath = true
		}
	}
	if !didGoPath {
		env = append(env, "GOPATH="+s.GOPATH)
	}
	cmd.Env = env

	// write up input to src
	cmd.Stdin = bytes.NewBuffer(src)

	// FIXME: this is not so great in case of errors, but simple and workable for the moment
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing %q with data from %q: %v  (combined output: %q)", cmdPath, filename, err, b)
	}

	return b, nil
}

type Generator interface {
	Generate(s *Settings, name string, args ...string) error
}

type GeneratorFunc func(s *Settings, name string, args ...string) error

func (f GeneratorFunc) Generate(s *Settings, name string, args ...string) error {
	return f(s, name, args...)
}

// globalMapGenerator acts as our registry
var globalMapGenerator = make(MapGenerator)

func MustRegister(name string, g Generator) {
	globalMapGenerator[name] = g
}

func GetRegistryMapGenerator() MapGenerator {
	return globalMapGenerator
}

// MapGenerator implements Generator by delegated based on name.
type MapGenerator map[string]Generator

func (g MapGenerator) Generate(s *Settings, name string, args ...string) error {

	theGen := g[name]
	if theGen != nil {
		return theGen.Generate(s, name, args...)
	}

	return fmt.Errorf("no generator %q found", name)
}

func andOneFile(s *Settings, targetFileIn string) (targetFile string, retdata map[string]interface{}, reterr error) {

	var err error

	targetFile, err = s.RelativeToGOPATH(targetFileIn)
	if err != nil {
		return "", nil, err
	}

	// detect existing package name, if any
	targetDir, _ := path.Split(targetFile)
	packageName, err := DetectDirPackage(filepath.Join(s.GOPATH, targetDir))
	if err == ErrNoPackageNameFound {
		targetFileSlash := filepath.ToSlash(targetFile)
		targetFileSlashParts := strings.Split(targetFileSlash, "/")
		packageName = targetFileSlashParts[len(targetFileSlashParts)-2]
	} else if err != nil {
		return "", nil, err
	}

	data := map[string]interface{}{
		"PackageName": packageName,
	}

	return targetFile, data, nil

}

// ParsePFlagsAndOneFile will parse arguments against a pflag set of flags.
func ParsePFlagsAndOneFile(s *Settings, fset *pflag.FlagSet, args []string) (targetFile string, retdata map[string]interface{}, reterr error) {
	err := fset.Parse(args)
	if err != nil {
		return "", nil, err
	}
	targetFile = fset.Arg(0)

	if targetFile == "" {
		return "", nil, fmt.Errorf("no target file specified")
	}

	if fset.NArg() > 1 {
		return "", nil, fmt.Errorf("more than one file argument specified")
	}

	return andOneFile(s, targetFile)
}

// ParseFlagsAndOneFile is deprecated, ParsePFlagsAndOneFile is recommended instead.
func ParseFlagsAndOneFile(s *Settings, fset *flag.FlagSet, args []string) (targetFile string, retdata map[string]interface{}, reterr error) {

	err := fset.Parse(args)
	if err != nil {
		return "", nil, err
	}
	targetFile = fset.Arg(0)

	if targetFile == "" {
		return "", nil, fmt.Errorf("no target file specified")
	}

	if fset.NArg() > 1 {
		return "", nil, fmt.Errorf("more than one file argument specified")
	}

	return andOneFile(s, targetFile)

	// targetFile, err = s.RelativeToGOPATH(targetFile)
	// if err != nil {
	// 	return "", nil, err
	// }

	// // detect existing package name, if any
	// targetDir, _ := path.Split(targetFile)
	// packageName, err := DetectDirPackage(filepath.Join(s.GOPATH, targetDir))
	// if err == ErrNoPackageNameFound {
	// 	targetFileSlash := filepath.ToSlash(targetFile)
	// 	targetFileSlashParts := strings.Split(targetFileSlash, "/")
	// 	packageName = targetFileSlashParts[len(targetFileSlashParts)-2]
	// } else if err != nil {
	// 	return "", nil, err
	// }

	// data := map[string]interface{}{
	// 	"PackageName": packageName,
	// }

	// return targetFile, data, nil
}

// GoSrcReplace performs a regexp replace on a file inline.
// Useful for adding things to an existing file.  The filePath
// is relative to and joined with s.GOPATH.
func GoSrcReplace(s *Settings, filePath string, pattern *regexp.Regexp, repl func(string) string) error {

	fullPath := filepath.Join(s.GOPATH, filePath)

	b, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// FIXME: this should somehow error if nothing was replaced...
	outs := pattern.ReplaceAllStringFunc(string(b), repl)

	// gofmt/goimports before writing back
	outBFmt, err := s.FormatGoCode(filePath, []byte(outs))

	return ioutil.WriteFile(fullPath, outBFmt, 0644)
}

func OutputGoSrcTemplate(s *Settings, data map[string]interface{}, targetFile string, tmplSrc string, debug bool) error {

	var buf bytes.Buffer
	var err error

	t := template.New("_out_")
	t = t.Funcs(template.FuncMap(map[string]interface{}{
		"bq": func(s string) string {
			return "`" + s + "`"
		},
		"plural": func(s string) string {
			return inflection.Plural(s)
		},
	}))

	t, err = t.Parse(tmplSrc)
	if err != nil {
		return err
	}

	err = t.Execute(&buf, data)
	if err != nil {
		return err
	}

	if debug {
		log.Printf("Generated program before formatting:\n%s", buf.Bytes())
	}

	b, err := s.FormatGoCode(targetFile, buf.Bytes())
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(s.GOPATH, targetFile), b, 0644)
	if err != nil {
		return err
	}

	return nil
}

// NameSnakeToCamel converts "some-name" to SomeName.
// Intended for deducing struct names from file names.
// You can optionally give a list of prefixes and suffixes to trim
// from the string before starting.  If nil is provided, no prefixes are
// trimmed but ".go" is trimmed as a suffix.
func NameSnakeToCamel(s string, trimPrefixes []string, trimSuffixes []string) string {

	if trimSuffixes == nil {
		trimSuffixes = []string{".go"}
	}

	for _, tp := range trimPrefixes {
		s = strings.TrimPrefix(s, tp)
	}

	for _, ts := range trimSuffixes {
		s = strings.TrimSuffix(s, ts)
	}

	parts := strings.Split(s, "-")

	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = string(unicode.ToUpper(rune(parts[i][0]))) + parts[i][1:]
		}
	}

	return strings.Join(parts, "")
}

var ErrNoPackageNameFound = fmt.Errorf("no package name found")

// DetectDirPackage will look at a Go directory and return the package name for existing files in it.
func DetectDirPackage(dirPath string) (string, error) {

	fileSet := token.NewFileSet()
	pmap, err := parser.ParseDir(fileSet, dirPath, nil, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	for k := range pmap {
		return k, nil
	}
	return "", ErrNoPackageNameFound
}
