package fsutil

import (
	"fmt"
	"path/filepath"
)

// MustAbs calls filepath.Abs() but panics on error instead of returning it.
func MustAbs(p string) string {
	ret, err := filepath.Abs(p)
	if err != nil {
		panic(fmt.Errorf("Error trying to get absolute path on %q: %v", p, err))
	}
	return ret
}
