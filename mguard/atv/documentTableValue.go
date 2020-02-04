package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// documentTableValue represents a table value in an ATV document.
type documentTableValue struct {
	Pos        lexer.Position
	Attributes dictionary          `"{" @@*`
	Rows       []*documentTableRow `@@* "}"`
}

// Dupe returns a deep copy of the table value.
func (table *documentTableValue) Dupe() *documentTableValue {

	if table == nil {
		return nil
	}

	var rowsCopy []*documentTableRow
	for _, row := range table.Rows {
		rowsCopy = append(rowsCopy, row.Dupe())
	}

	return &documentTableValue{
		Attributes: table.Attributes,
		Rows:       rowsCopy,
	}
}

// GetRowReferences returns all row references recursively.
func (table *documentTableValue) GetRowReferences() []RowRef {

	if table != nil {
		var allRowRefs []RowRef
		for _, row := range table.Rows {
			allRowRefs = append(allRowRefs, row.GetRowReferences()...)
		}
		return allRowRefs
	}
	return []RowRef{}
}

// GetRowIDs returns all row ids recursively.
func (table *documentTableValue) GetRowIDs() []RowID {

	if table != nil {
		var allRowIDs []RowID
		for _, row := range table.Rows {
			allRowIDs = append(allRowIDs, row.GetRowIDs()...)
		}
		return allRowIDs
	}
	return []RowID{}
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (table *documentTableValue) WriteDocumentPart(writer *strings.Builder, indent int) error {

	// write opening brace for table
	_, err := writer.WriteString("{\n")
	if err != nil {
		return err
	}

	// write key-value-pairs forming the attributes
	for _, item := range table.Attributes {
		line := fmt.Sprintf("%s%s = \"%s\"\n", spacer(indent+1), item.Key, item.Value)
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
func (table *documentTableValue) String() string {

	if table == nil {
		return "<nil>"
	}

	builder := strings.Builder{}
	table.WriteDocumentPart(&builder, 0)
	return strings.TrimSpace(builder.String())
}
