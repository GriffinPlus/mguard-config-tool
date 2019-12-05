package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// Setting represents a setting node in an ATV document.
type Setting struct {
	Pos         lexer.Position
	Name        string  `@Ident "="`
	SimpleValue *string `( @String`
	TableValue  *Table  `| @@`
	RowRef      *RowRef `| @@ )`
}

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

// RowRef represents a row reference in an ATV document.
type RowRef struct {
	Pos   lexer.Position
	RowID string `"{" "rowref" "=" @String "}"`
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (setting *Setting) WriteDocumentPart(writer *strings.Builder, indent int) error {
	if setting == nil {
		return nil
	}

	if setting.SimpleValue != nil {

		line := fmt.Sprintf("%s%s = \"%s\"\n", spacer(indent), setting.Name, *setting.SimpleValue)
		_, err := writer.WriteString(line)
		return err

	} else if setting.TableValue != nil {

		// write setting name and opening brace for table
		line := fmt.Sprintf("%s%s = {\n", spacer(indent), setting.Name)
		_, err := writer.WriteString(line)
		if err != nil {
			return err
		}

		// write UUID of the setting, if available
		if setting.TableValue.UUID != nil {
			line := fmt.Sprintf("%suuid = \"%s\"\n", spacer(indent+1), *setting.TableValue.UUID)
			_, err := writer.WriteString(line)
			if err != nil {
				return err
			}
		}

		// write table rows
		for _, row := range setting.TableValue.Rows {

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
		line = fmt.Sprintf("%s}\n", spacer(indent))
		_, err = writer.WriteString(line)
		return err

	} else if setting.RowRef != nil {

		// write reference to a table row
		line := fmt.Sprintf(
			"%s%s = {\n%srowref = \"%s\"\n%s}\n",
			spacer(indent), setting.Name,
			spacer(indent+1), setting.RowRef.RowID,
			spacer(indent))
		_, err := writer.WriteString(line)
		return err

	} else {

		panic("Unhandled value type")

	}

}
