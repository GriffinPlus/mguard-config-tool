package atv

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// documentPragma represents a pragma in an ATV configuration document.
type documentPragma struct {
	Pos   lexer.Position
	Name  string
	Value string
}

// Dupe returns a copy of the document node.
func (pragma *documentPragma) Dupe() *documentPragma {

	if pragma == nil {
		return nil
	}

	return &documentPragma{
		Name:  pragma.Name,
		Value: pragma.Value,
	}
}

var pragmaSplitterRegex = regexp.MustCompile(`#(\w+)(?:\s+(.*))?$`)

// Capture initializes the pragma object when the parser encounters a pragma token in an ATV document.
func (pragma *documentPragma) Capture(values []string) error {
	match := pragmaSplitterRegex.FindStringSubmatch(values[0])
	if len(match) == 0 {
		return fmt.Errorf("Splitting pragma failed")
	}
	pragma.Name = match[1]
	pragma.Value = match[2]
	return nil
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (pragma *documentPragma) WriteDocumentPart(writer *strings.Builder, indent int) error {
	line := fmt.Sprintf("%s#%s %s\n", spacer(indent), pragma.Name, pragma.Value)
	_, err := writer.WriteString(line)
	return err
}
