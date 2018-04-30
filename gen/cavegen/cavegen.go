package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/gocaveman/caveman/gen"
)

func main() {

	args := os.Args[1:]

	g := gen.GetRegistryMapGenerator()

	if len(args) == 0 || g[args[0]] == nil {
		fmt.Printf("Usage: cavegen [generator] [args...]\n\n")
		fmt.Printf("  Generators:\n")
		gnames := make([]string, 0, len(g))
		for name := range g {
			gnames = append(gnames, name)
		}
		sort.Strings(gnames)
		for _, name := range gnames {
			fmt.Printf("    %s\n", name)
		}

		// fmt.Printf("\nTo get help on a specific generator: cavegen help [generator]\n\n")
		os.Exit(1)
		return
	}

	// if args[0] == "help" {
	// 	panic("help not implemented!")
	// 	os.Exit(1)
	// 	return
	// }

	name := args[0]

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Printf("GOPATH not set, cannot continue")
		os.Exit(1)
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	s := &gen.Settings{
		GOPATH:  gopath,
		WorkDir: wd,
	}

	err = g.Generate(s, name, args[1:]...)
	if err != nil {
		log.Printf("Generate() produced error: %v", err)
		os.Exit(255)
		return
	}

}
