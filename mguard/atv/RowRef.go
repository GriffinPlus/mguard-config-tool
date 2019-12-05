package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// RowRef represents a row reference in an ATV document.
type RowRef struct {
	Pos   lexer.Position
	RowID string `"{" "rowref" "=" @String "}"`
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (rowref *RowRef) WriteDocumentPart(writer *strings.Builder, indent int) error {

	line := fmt.Sprintf(
		"{\n%srowref = \"%s\"\n%s}\n",
		spacer(indent+1), rowref.RowID, spacer(indent))
	_, err := writer.WriteString(line)
	return err
}
