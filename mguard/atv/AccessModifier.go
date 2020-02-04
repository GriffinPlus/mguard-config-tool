package atv

import "fmt"

// AccessModifier influences the access of a setting in an ATV document.
type AccessModifier int

const (
	// MustNotOverwrite indicates that a setting must not be overwritten.
	MustNotOverwrite AccessModifier = iota

	// MayOverwrite indicates that a setting may be overwritten.
	MayOverwrite

	// MustOverwrite indicates that a setting must be overwritten.
	MustOverwrite

	// MayAppend indicates that a setting can be extended by appending additional rows (table values only).
	MayAppend

	// DefaultAccessModifier is the access modifier that applys, if no access modifier is specified (for internal use only).
	DefaultAccessModifier
)

var accessModifierMapping = []string{
	"must-not-overwrite", // MustNotOverwrite
	"may-overwrite",      // MayOverwrite
	"must-overwrite",     // MustOverwrite
	"may-append",         // MayAppend
	"may-overwrite",      // DefaultAccessModifier
}

// String returns the string representation of the access modifier.
func (access AccessModifier) String() string {
	return accessModifierMapping[access]
}

// ParseAccessModifier parses the specified string as an access modifier.
func ParseAccessModifier(s string) (AccessModifier, error) {
	for i, item := range accessModifierMapping {
		if item == s {
			return AccessModifier(i), nil
		}
	}
	return DefaultAccessModifier, fmt.Errorf("'%s' is not a valid access modifier", s)
}
