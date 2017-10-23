// +build ignore

package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/shurcooL/vfsgen"
	"github.com/spf13/afero"
)

func main() {

	wd, _ := os.Getwd()
	includesDir := filepath.Join(wd)
	includesFS := afero.NewHttpFs(afero.NewRegexpFs(afero.NewBasePathFs(afero.NewOsFs(), includesDir),
		regexp.MustCompile(`\.(gohtml|html|css|js)$`)))

	err := vfsgen.Generate(includesFS, vfsgen.Options{
		PackageName:  "demotheme",
		BuildTags:    "!dev",
		VariableName: "assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
