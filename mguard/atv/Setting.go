package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// Setting represents a setting node in an ATV document.
type Setting struct {
	Pos         lexer.Position
	Name        string       `@Ident "="`
	SimpleValue *SimpleValue `( @@`
	TableValue  *TableValue  `| @@`
	RowRef      *RowRef      `| @@ )`
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (setting *Setting) WriteDocumentPart(writer *strings.Builder, indent int) error {

	// write the name of the setting
	line := fmt.Sprintf("%s%s = ", spacer(indent), setting.Name)
	_, err := writer.WriteString(line)
	if err != nil {
		return err
	}

	// write the setting value
	parts := []DocumentWriter{setting.SimpleValue, setting.TableValue, setting.RowRef}
	for _, part := range parts {

		if !isNil(part) {
			err := part.WriteDocumentPart(writer, indent)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
