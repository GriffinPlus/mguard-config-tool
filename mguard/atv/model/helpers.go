package model

import (
	"fmt"
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

// charsToQuote contains characters that are quoted in string in ATV documents.
const charsToQuote = `"\`

// quote quotes the specified string in the ATV specific fashion.
func quote(s string) string {
	quoted := strings.Builder{}
	quoted.WriteRune('"')
	for _, c := range s {
		if strings.ContainsRune(charsToQuote, c) {
			quoted.WriteRune('\\')
		}
		quoted.WriteRune(c)
	}
	quoted.WriteRune('"')
	return quoted.String()
}

// unquote unquotes the specified string in the ATV specific fashion.
func unquote(s string) (string, error) {

	// check whether the string contains the expected embracing quotes
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", fmt.Errorf("String is not properly quoted")
	}

	// strip embracing quotes
	s = s[1 : len(s)-1]

	unquoted := strings.Builder{}
	unquotePending := false
	for _, c := range s {
		if !unquotePending && c == '\\' {
			unquotePending = true
			continue
		}

		if unquotePending && !strings.ContainsRune(charsToQuote, c) {
			unquoted.WriteRune('\\')
		}

		unquoted.WriteRune(c)
		unquotePending = false
	}

	if unquotePending {
		return "", fmt.Errorf("Unpaired backslash found when unquoting string (%s)", s)
	}

	return unquoted.String(), nil
}
