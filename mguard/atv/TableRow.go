package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// TableRow represents a table row in an ATV document.
type TableRow struct {
	Pos   lexer.Position
	RowID *RowID     `"{" ( "{" "rid" "=" @String "}" )?`
	Items []*Setting `@@* "}"`
}

// Dupe returns a deep copy of the table row.
func (row *TableRow) Dupe() *TableRow {

	if row == nil {
		return nil
	}

	var itemsCopy []*Setting
	for _, setting := range row.Items {
		itemsCopy = append(itemsCopy, setting.Dupe())
	}

	return &TableRow{
		RowID: row.RowID,
		Items: itemsCopy,
	}
}

// GetRowReferences returns all row references recursively.
func (row *TableRow) GetRowReferences() []*RowRef {

	if row == nil {
		return []*RowRef{}
	}

	var allRowRefs []*RowRef
	for _, item := range row.Items {
		allRowRefs = append(allRowRefs, item.GetRowReferences()...)
	}

	return allRowRefs
}

// GetRowIDs returns all row ids recursively.
func (row *TableRow) GetRowIDs() []RowID {

	if row == nil {
		return []RowID{}
	}

	var allRowIDs []RowID
	if row.RowID != nil {
		allRowIDs = append(allRowIDs, *row.RowID)
	}

	for _, item := range row.Items {
		allRowIDs = append(allRowIDs, item.GetRowIDs()...)
	}

	return allRowIDs
}

// HasID checks whether the row has a row id.
func (row *TableRow) HasID() bool {
	return row != nil && row.RowID != nil
}

// HasSameID checks whether the current row and the specified one has the same row id.
func (row *TableRow) HasSameID(other *TableRow) bool {
	return row != nil && other != nil && row.RowID != nil && other.RowID != nil && row.RowID == other.RowID
}

// String returns the table row as a string.
func (row *TableRow) String() string {

	if row == nil {
		return "<nil>"
	}

	builder := strings.Builder{}
	row.WriteDocumentPart(&builder, 0)
	return strings.TrimSuffix(builder.String(), "\n")
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (row *TableRow) WriteDocumentPart(writer *strings.Builder, indent int) error {

	// write opening brace of the table row
	line := fmt.Sprintf("%s{\n", spacer(indent))
	_, err := writer.WriteString(line)
	if err != nil {
		return err
	}

	// write row id, if available
	if row.RowID != nil {
		line := fmt.Sprintf("%s{ rid = \"%s\" }\n", spacer(indent+1), *row.RowID)
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	// write settings in the table row
	for _, item := range row.Items {
		err = item.WriteDocumentPart(writer, indent+1)
		if err != nil {
			return err
		}
	}

	// write closing brace of the table row
	line = fmt.Sprintf("%s}\n", spacer(indent))
	_, err = writer.WriteString(line)
	if err != nil {
		return err
	}

	return nil
}
