package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// Table represents a table node in an ATV document.
type Table struct {
	Pos  lexer.Position
	UUID *string     `"{" ( "uuid" "=" @String )?`
	Rows []*TableRow `@@* "}"`
}

// TableRow represents a table row in an ATV document.
type TableRow struct {
	Pos   lexer.Position
	RowID *string    `"{" ( "{" "rid" "=" @String "}" )?`
	Items []*Setting `@@* "}"`
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (table *Table) WriteDocumentPart(writer *strings.Builder, indent int) error {

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

		// write opening brace of the table row
		line := fmt.Sprintf("%s{\n", spacer(indent+1))
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}

		// write row id, if available
		if row.RowID != nil {
			line := fmt.Sprintf("%s{ rid = \"%s\" }\n", spacer(indent+2), *row.RowID)
			_, err := writer.WriteString(line)
			if err != nil {
				return err
			}
		}

		// write settings in the table row
		for _, item := range row.Items {
			err = item.WriteDocumentPart(writer, indent+2)
			if err != nil {
				return err
			}
		}

		// write closing brace of the table row
		line = fmt.Sprintf("%s}\n", spacer(indent+1))
		_, err = writer.WriteString(line)
		if err != nil {
			return err
		}
	}

	// write closing brace of the table
	line := fmt.Sprintf("%s}\n", spacer(indent))
	_, err = writer.WriteString(line)
	return err
}