package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// SimpleValue represents a simple value in an ATV document.
type SimpleValue struct {
	Pos   lexer.Position
	Value string `@String`
}

// Dupe returns a copy of the value.
func (value *SimpleValue) Dupe() *SimpleValue {

	if value == nil {
		return nil
	}

	return &SimpleValue{Value: value.Value}
}

// GetRowReferences returns all row references recursively.
func (value *SimpleValue) GetRowReferences() []*RowRef {
	return []*RowRef{}
}

// GetRowIDs returns all row ids recursively.
func (value *SimpleValue) GetRowIDs() []RowID {
	return []RowID{}
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (value *SimpleValue) WriteDocumentPart(writer *strings.Builder, indent int) error {
	line := fmt.Sprintf("%s\n", quote(value.Value))
	_, err := writer.WriteString(line)
	return err
}

// String returns the simple value as a string.
func (value *SimpleValue) String() string {

	if value == nil {
		return "<nil>"
	}

	return value.Value
}
