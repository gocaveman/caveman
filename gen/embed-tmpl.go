package gen

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

func init() {
	globalMapGenerator["embed-tmpl"] = GeneratorFunc(func(s *Settings, name string, args ...string) error {

		fset := pflag.NewFlagSet("gen", pflag.ContinueOnError)
		targetFile, data, err := ParsePFlagsAndOneFile(s, fset, args)
		if err != nil {
			return err
		}

		targetDir, targetFileName := filepath.Split(targetFile)
		if path.Ext(targetFileName) != ".go" {
			return fmt.Errorf("target file name (%q) must end with .go", targetFileName)
		}
		_ = targetDir

		targetGogenFileName := strings.Replace(targetFileName, ".go", "-gogen.go", 1)
		data["TargetGogenFileName"] = targetGogenFileName
		targetGogenFile := filepath.Join(targetDir, targetGogenFileName)

		targetDataFileName := strings.Replace(targetFileName, ".go", "-data.go", 1)
		data["TargetDataFileName"] = targetDataFileName
		targetDataDevFileName := strings.Replace(targetFileName, ".go", "-data-dev.go", 1)
		data["TargetDataDevFileName"] = targetDataDevFileName

		err = OutputGoSrcTemplate(s, data, targetFile, `
package {{.PackageName}}

import (
	"net/http"
	"path"

	"github.com/gocaveman/caveman/filesystem/fsutil"
	"github.com/gocaveman/caveman/tmpl"
	"github.com/gocaveman/caveman/tmpl/tmplregistry"
)

//go:generate go run {{.TargetGogenFileName}}

func init() {

	// TODO: static stuff

	tmplregistry.MustRegister(tmplregistry.SeqTheme, "{{.PackageName}}", NewTmplStore())

}

func NewViewsFS() http.FileSystem {
	return fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return EmbeddedAssets.Open("/views" + path.Clean("/"+name))
	})
}

func NewIncludesFS() http.FileSystem {
	return fsutil.NewHTTPFuncFS(func(name string) (http.File, error) {
		return EmbeddedAssets.Open("/includes" + path.Clean("/"+name))
	})
}

func NewTmplStore() tmpl.Store {
	return &tmpl.HFSStore{
		FileSystems: map[string]http.FileSystem{
			tmpl.ViewsCategory:    NewViewsFS(),
			tmpl.IncludesCategory: NewIncludesFS(),
		},
	}
}
`, false)
		if err != nil {
			return err
		}

		err = OutputGoSrcTemplate(s, data, targetGogenFile, `// +build ignore

package main

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/shurcooL/vfsgen"
	"github.com/spf13/afero"
)

func main() {

	packageName := "{{.PackageName}}"

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	base := afero.NewBasePathFs(afero.NewOsFs(), wd)
	filteredAFS := afero.NewRegexpFs(base, regexp.MustCompile("\\.(gohtml|html|css|js|jpg|png|gif|svg)$"))
	filteredHFS := afero.NewHttpFs(filteredAFS)
	err = vfsgen.Generate(filteredHFS, vfsgen.Options{
		PackageName:  packageName,
		BuildTags:    "!dev",
		VariableName: "EmbeddedAssets",
		Filename:     "{{.TargetDataFileName}}",
	})
	if err != nil {
		log.Fatalln(err)
	}

	devContents := `+"`"+`// +build dev

package __PACKAGE__

// CODE GENERATED FILE, DO NOT EDIT!

import (
	"net/http"
)

// EmbeddedAssets dev version loads from local filesystem.
var EmbeddedAssets = func() http.FileSystem {
	return http.Dir("__LOCALPATH__")
}()
`+"`"+`

	devStr := strings.Replace(string(devContents), "__PACKAGE__", packageName, -1)
	devStr = strings.Replace(devStr, "__LOCALPATH__", wd, -1)

	err = ioutil.WriteFile("{{.TargetDataDevFileName}}", []byte(devStr), 0644)
	if err != nil {
		log.Fatal(err)
	}

}

`, false)
		if err != nil {
			return err
		}

		return nil
	})
}
