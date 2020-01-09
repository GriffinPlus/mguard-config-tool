package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// TableValue represents a table value in an ATV document.
type TableValue struct {
	Pos  lexer.Position
	UUID *string     `"{" ( "uuid" "=" @String )?`
	Rows []*TableRow `@@* "}"`
}

// Dupe returns a deep copy of the table value.
func (table *TableValue) Dupe() *TableValue {

	if table == nil {
		return nil
	}

	var rowsCopy []*TableRow
	for _, row := range table.Rows {
		rowsCopy = append(rowsCopy, row.Dupe())
	}

	return &TableValue{
		UUID: table.UUID,
		Rows: rowsCopy,
	}
}

// GetRowReferences returns all row references recursively.
func (table *TableValue) GetRowReferences() []*RowRef {

	if table == nil {
		return []*RowRef{}
	}

	var allRowRefs []*RowRef
	for _, row := range table.Rows {
		allRowRefs = append(allRowRefs, row.GetRowReferences()...)
	}

	return allRowRefs
}

// GetRowIDs returns all row ids recursively.
func (table *TableValue) GetRowIDs() []RowID {

	if table == nil {
		return []RowID{}
	}

	var allRowIDs []RowID
	for _, row := range table.Rows {
		allRowIDs = append(allRowIDs, row.GetRowIDs()...)
	}

	return allRowIDs
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (table *TableValue) WriteDocumentPart(writer *strings.Builder, indent int) error {

	// write opening brace for table
	_, err := writer.WriteString("{\n")
	if err != nil {
		return err
	}

	// write UUID of the setting, if available
	if table.UUID != nil {
		line := fmt.Sprintf("%suuid = \"%s\"\n", spacer(indent+1), *table.UUID)
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	// write table rows
	for _, row := range table.Rows {
		row.WriteDocumentPart(writer, indent+1)
	}

	// write closing brace of the table
	line := fmt.Sprintf("%s}\n", spacer(indent))
	_, err = writer.WriteString(line)
	return err
}

// String returns the table value as a string.
func (table *TableValue) String() string {

	if table == nil {
		return "<nil>"
	}

	builder := strings.Builder{}
	table.WriteDocumentPart(&builder, 0)
	return strings.TrimSuffix(builder.String(), "\n")
}
