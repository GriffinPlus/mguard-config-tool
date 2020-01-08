package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// Setting represents a setting node in an ATV document.
type Setting struct {
	Pos          lexer.Position
	Name         string        `@Ident "="`
	SimpleValue  *SimpleValue  `( @@`
	TableValue   *TableValue   `| @@`
	RowRef       *RowRef       `| @@`
	ComplexValue *ComplexValue `| @@ )`
}

// Dupe returns a deep copy of the setting.
func (setting *Setting) Dupe() *Setting {
	var copy = &Setting{Name: setting.Name}
	if setting.SimpleValue != nil {
		copy.SimpleValue = setting.SimpleValue.Dupe()
	} else if setting.ComplexValue != nil {
		copy.ComplexValue = setting.ComplexValue.Dupe()
	} else if setting.TableValue != nil {
		copy.TableValue = setting.TableValue.Dupe()
	} else if setting.RowRef != nil {
		copy.RowRef = setting.RowRef.Dupe()
	} else {
		panic("Unhandled setting value")
	}
	return copy
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
	parts := []DocumentWriter{setting.SimpleValue, setting.ComplexValue, setting.RowRef, setting.TableValue}
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
