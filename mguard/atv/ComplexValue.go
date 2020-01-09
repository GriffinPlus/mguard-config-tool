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

	if value == nil {
		return nil
	}

	var itemsCopy []*Setting
	for _, setting := range value.Items {
		itemsCopy = append(itemsCopy, setting.Dupe())
	}

	return &ComplexValue{
		UUID:  value.UUID,
		Items: itemsCopy,
	}
}

// GetRowReferences returns all row references recursively.
func (value *ComplexValue) GetRowReferences() []*RowRef {

	if value == nil {
		return []*RowRef{}
	}

	var allRowRefs []*RowRef
	for _, item := range value.Items {
		allRowRefs = append(allRowRefs, item.GetRowReferences()...)
	}

	return allRowRefs
}

// GetRowIDs returns all row ids recursively.
func (value *ComplexValue) GetRowIDs() []RowID {

	if value == nil {
		return []RowID{}
	}

	var allRowIDs []RowID
	for _, item := range value.Items {
		allRowIDs = append(allRowIDs, item.GetRowIDs()...)
	}

	return allRowIDs
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (value *ComplexValue) WriteDocumentPart(writer *strings.Builder, indent int) error {

	// write opening brace of the complex type
	_, err := writer.WriteString("{\n")
	if err != nil {
		return err
	}

	// write UUID of the setting, if available
	if value.UUID != nil {
		line := fmt.Sprintf("%suuid = \"%s\"\n", spacer(indent+1), *value.UUID)
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	// write settings
	for _, item := range value.Items {
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

// String returns the complex value as a string.
func (value *ComplexValue) String() string {
	builder := strings.Builder{}
	value.WriteDocumentPart(&builder, 0)
	return strings.TrimSuffix(builder.String(), "\n")
}
