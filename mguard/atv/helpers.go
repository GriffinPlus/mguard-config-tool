package atv

import "strings"

// spacer generates a whitespace string that can be used for indenting lines up to the specified level.
func spacer(level int) string {
	return strings.Repeat("  ", level)
}
