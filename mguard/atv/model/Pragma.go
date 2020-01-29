package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// Pragma represents a pragma in an ATV configuration document.
type Pragma struct {
	Pos   lexer.Position
	Name  string
	Value string
}

// Dupe returns a copy of the document node.
func (pragma *Pragma) Dupe() *Pragma {

	if pragma == nil {
		return nil
	}

	return &Pragma{
		Name:  pragma.Name,
		Value: pragma.Value,
	}
}

var pragmaSplitterRegex = regexp.MustCompile(`#(\w+)(?:\s+(.*))?$`)

// Capture initializes the pragma object when the parser encounters a pragma token in an ATV document.
func (pragma *Pragma) Capture(values []string) error {
	match := pragmaSplitterRegex.FindStringSubmatch(values[0])
	if len(match) == 0 {
		return fmt.Errorf("Splitting pragma failed")
	}
	pragma.Name = match[1]
	pragma.Value = match[2]
	return nil
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (pragma *Pragma) WriteDocumentPart(writer *strings.Builder, indent int) error {
	line := fmt.Sprintf("%s#%s %s\n", spacer(indent), pragma.Name, pragma.Value)
	_, err := writer.WriteString(line)
	return err
}
