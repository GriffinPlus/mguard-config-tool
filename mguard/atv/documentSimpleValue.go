package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// documentSimpleValue represents a simple value in an ATV document.
type documentSimpleValue struct {
	Pos   lexer.Position
	Value string `@String`
}

// Dupe returns a copy of the value.
func (value *documentSimpleValue) Dupe() *documentSimpleValue {

	if value == nil {
		return nil
	}

	return &documentSimpleValue{Value: value.Value}
}

// GetRowReferences returns all row references recursively.
func (value *documentSimpleValue) GetRowReferences() []RowRef {
	return []RowRef{}
}

// GetRowIDs returns all row ids recursively.
func (value *documentSimpleValue) GetRowIDs() []RowID {
	return []RowID{}
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (value *documentSimpleValue) WriteDocumentPart(writer *strings.Builder, indent int) error {
	line := fmt.Sprintf("%s\n", quote(value.Value))
	_, err := writer.WriteString(line)
	return err
}

// String returns the simple value as a string.
func (value *documentSimpleValue) String() string {

	if value == nil {
		return "<nil>"
	}

	return strings.TrimSpace(value.Value)
}
