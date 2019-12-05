package atv

import (
	"reflect"
	"strings"
)

// spacer generates a whitespace string that can be used for indenting lines up to the specified level.
func spacer(level int) string {
	return strings.Repeat("  ", level)
}

// isNil checks whether the value of the specified interface is nil.
func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}
