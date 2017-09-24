package caveman

import (
	"errors"
	"fmt"
	"regexp"
)

var VALID_TOKEN_REGEXP = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func CheckValidToken(t string) error {
	if !VALID_TOKEN_REGEXP.MatchString(t) {
		return fmt.Errorf("invalid token, must contain only letters, digits or underscores and start with a letter or underscore")
	}
	return nil
}

var ErrNotFound = errors.New("not found")
