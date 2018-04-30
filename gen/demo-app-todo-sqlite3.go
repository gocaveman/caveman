package gen

import (
	"path/filepath"

	"github.com/spf13/pflag"
)

func init() {
	globalMapGenerator["demo-app-todo-sqlite3"] = GeneratorFunc(func(s *Settings, name string, args ...string) error {

		// generate our demo application by calling the other generators

		fset := pflag.NewFlagSet("gen", pflag.ContinueOnError)
		targetFile, _, err := ParsePFlagsAndOneFile(s, fset, args)
		if err != nil {
			return err
		}

		targetDir, _ := filepath.Split(targetFile)

		// FIXME: We probably should generate this with whatever package structure
		// we are going to recommend for projects, and it's probably not the
		// "everything in main" approach.

		err = globalMapGenerator.Generate(s, "main-rest-sqlite3", targetFile)
		if err != nil {
			return err
		}
		err = globalMapGenerator.Generate(s, "model-sample-todo-item", filepath.Join(targetDir, "model-todo-item.go"))
		if err != nil {
			return err
		}
		err = globalMapGenerator.Generate(s, "store", filepath.Join(targetDir, "store.go"))
		if err != nil {
			return err
		}
		err = globalMapGenerator.Generate(s, "store-crud", filepath.Join(targetDir, "store-todo-item.go"))
		if err != nil {
			return err
		}
		err = globalMapGenerator.Generate(s, "ctrl-crud", filepath.Join(targetDir, "ctrl-todo-item.go"))
		if err != nil {
			return err
		}

		return nil
	})
}
