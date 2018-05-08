// +build ignore

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

	packageName := "paleolithic"

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
		Filename:     "embed-data.go",
	})
	if err != nil {
		log.Fatalln(err)
	}

	devContents := `// +build dev

package __PACKAGE__

// CODE GENERATED FILE, DO NOT EDIT!

import (
	"net/http"
)

// EmbeddedAssets dev version loads from local filesystem.
var EmbeddedAssets = func() http.FileSystem {
	return http.Dir("__LOCALPATH__")
}()
`

	devStr := strings.Replace(string(devContents), "__PACKAGE__", packageName, -1)
	devStr = strings.Replace(devStr, "__LOCALPATH__", wd, -1)

	err = ioutil.WriteFile("embed-data-dev.go", []byte(devStr), 0644)
	if err != nil {
		log.Fatal(err)
	}

}
