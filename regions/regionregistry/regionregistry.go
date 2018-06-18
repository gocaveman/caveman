// Registry of region defintions.
package regionregistry

import "github.com/gocaveman/caveman/regions"

var global regions.DefinitionList

// MustRegister adds regions to global registry
func MustRegister(def ...regions.Definition) {
	global = append(global, def...)
}

// Contents returns a copy of the global region definition registry.
func Contents() regions.DefinitionList {
	return global.SortedCopy()
}
