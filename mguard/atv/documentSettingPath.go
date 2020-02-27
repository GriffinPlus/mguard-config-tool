package atv

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// documentSettingPathToken represents a token in a setting path.
type documentSettingPathToken struct {
	name *string // set, if the token specifies a setting name.
	row  *int    // set, if the token specifies a row in a table
}

// documentSettingPath represents a parsed setting path.
type documentSettingPath []documentSettingPathToken

var settingNameRegex = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
var tableRowAccessRegex = regexp.MustCompile(`^[0-9]+$`)

// parseDocumentSettingPath parses the specified string and returns the corresponding tokens.
func parseDocumentSettingPath(s string) (documentSettingPath, error) {

	var tokens documentSettingPath
	settingPreceding := false
	for _, token := range strings.Split(s, ".") {

		if settingNameRegex.MatchString(token) {

			if settingPreceding {
				return nil, fmt.Errorf("Invalid path, '%s' not expected", token)
			}

			copy := token // workaround to avoid reusing iteration variable which would break the resulting tokens
			tokens = append(tokens, documentSettingPathToken{name: &copy})
			settingPreceding = true
			continue
		}

		if tableRowAccessRegex.MatchString(token) {

			if !settingPreceding {
				return nil, fmt.Errorf("Invalid path, '%s' not expected", token)
			}

			row, err := strconv.Atoi(token)
			if err != nil {
				return nil, err
			}

			tokens = append(tokens, documentSettingPathToken{row: &row})
			settingPreceding = false
			continue
		}

		return nil, fmt.Errorf("Invalid path, '%s' is not a valid path token", token)
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("Path is empty")
	}

	return tokens, nil
}

// String returns the string representation of the token.
func (token documentSettingPathToken) String() string {

	if token.name != nil {
		return *token.name
	}

	if token.row != nil {
		return fmt.Sprintf("%d", *token.row)
	}

	panic("Unhandled token type")
}

// String returns the string representation of the path.
func (path documentSettingPath) String() string {

	builder := strings.Builder{}
	for i, token := range path {
		if i > 0 {
			builder.WriteString(".")
		}
		builder.WriteString(token.String())
	}

	return builder.String()
}
