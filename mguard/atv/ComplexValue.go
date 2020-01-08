package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// ComplexValue represents a complex value in an ATV document.
type ComplexValue struct {
	Pos   lexer.Position
	UUID  *string    `"{" ( "uuid" "=" @String )?`
	Items []*Setting `@@* "}"`
}

// Dupe returns a copy of the value.
func (value *ComplexValue) Dupe() *ComplexValue {

	var itemsCopy []*Setting
	for _, setting := range value.Items {
		itemsCopy = append(itemsCopy, setting.Dupe())
	}

	return &ComplexValue{
		UUID:  value.UUID,
		Items: itemsCopy,
	}
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (complex *ComplexValue) WriteDocumentPart(writer *strings.Builder, indent int) error {

	// write opening brace of the complex type
	_, err := writer.WriteString("{\n")
	if err != nil {
		return err
	}

	// write UUID of the setting, if available
	if complex.UUID != nil {
		line := fmt.Sprintf("%suuid = \"%s\"\n", spacer(indent+1), *complex.UUID)
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	// write settings
	for _, item := range complex.Items {
		err = item.WriteDocumentPart(writer, indent+1)
		if err != nil {
			return err
		}
	}

	// write closing brace of the complex type
	line := fmt.Sprintf("%s}\n", spacer(indent))
	_, err = writer.WriteString(line)
	if err != nil {
		return err
	}

	return err
}
