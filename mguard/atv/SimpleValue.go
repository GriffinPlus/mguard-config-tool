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
	return &SimpleValue{Value: value.Value}
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (value *SimpleValue) WriteDocumentPart(writer *strings.Builder, indent int) error {
	line := fmt.Sprintf("%s\n", quote(value.Value))
	_, err := writer.WriteString(line)
	return err
}
