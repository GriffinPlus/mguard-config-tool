package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// documentValueWithMetadata represents a value with some metadata attached to it in an ATV document.
type documentValueWithMetadata struct {
	Pos  lexer.Position
	Data dictionary `"{" @@* "}"`
}

// Dupe returns a copy of the value.
func (value *documentValueWithMetadata) Dupe() *documentValueWithMetadata {

	if value == nil {
		return nil
	}

	return &documentValueWithMetadata{
		Data: value.Data,
	}
}

// GetRowReferences returns all row references recursively.
func (value *documentValueWithMetadata) GetRowReferences() []RowRef {

	if value != nil {
		var rowref string
		if value.Data.TryGet("rowref", &rowref) {
			return []RowRef{RowRef(rowref)}
		}
	}

	return []RowRef{}
}

// GetRowIDs returns all row ids recursively.
func (value *documentValueWithMetadata) GetRowIDs() []RowID {
	return []RowID{}
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (value *documentValueWithMetadata) WriteDocumentPart(writer *strings.Builder, indent int) error {

	// write opening brace of the complex type
	_, err := writer.WriteString("{\n")
	if err != nil {
		return err
	}

	// write key-value-pairs forming the value
	for _, item := range value.Data {
		line := fmt.Sprintf("%s%s = \"%s\"\n", spacer(indent+1), item.Key, item.Value)
		_, err := writer.WriteString(line)
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
func (value *documentValueWithMetadata) String() string {
	builder := strings.Builder{}
	value.WriteDocumentPart(&builder, 0)
	return strings.TrimSpace(builder.String())
}
