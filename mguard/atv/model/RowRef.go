package model

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

// Dupe returns a copy of the value.
func (rowref *RowRef) Dupe() *RowRef {

	if rowref == nil {
		return nil
	}

	return &RowRef{RowID: rowref.RowID}
}

// GetRowReferences returns all row references recursively.
func (rowref *RowRef) GetRowReferences() []*RowRef {

	if rowref == nil {
		return []*RowRef{}
	}

	return []*RowRef{rowref}
}

// GetRowIDs returns all row ids recursively.
func (rowref *RowRef) GetRowIDs() []RowID {
	return []RowID{}
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (rowref *RowRef) WriteDocumentPart(writer *strings.Builder, indent int) error {
	line := fmt.Sprintf(
		"{\n%srowref = \"%s\"\n%s}\n",
		spacer(indent+1), rowref.RowID, spacer(indent))
	_, err := writer.WriteString(line)
	return err
}

// String returns the rowref value as a string.
func (rowref *RowRef) String() string {

	if rowref == nil {
		return "<nil>"
	}

	builder := strings.Builder{}
	rowref.WriteDocumentPart(&builder, 0)
	return strings.TrimSuffix(builder.String(), "\n")
}
