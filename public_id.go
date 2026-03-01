package bpid

import (
	"fmt"
	"regexp"
)

// PublicID is the interface that ID definition types must implement.
// It provides the string prefix for the ID type.
//
// Implementations are structs whose exported fields carry the ID's data:
//
//	type UserID struct{ OrgID, UserSeq int64 }
//	func (UserID) Prefix() string { return "user" }
type PublicID interface {
	Prefix() string
}

// prefixRegexp validates that a prefix contains only lowercase alphanumeric,
// hyphens, and underscores.
var prefixRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// validatePrefix checks that a prefix matches the allowed pattern.
func validatePrefix(prefix string) error {
	if !prefixRegexp.MatchString(prefix) {
		return fmt.Errorf("%w: got %q", ErrInvalidPrefix, prefix)
	}
	return nil
}
