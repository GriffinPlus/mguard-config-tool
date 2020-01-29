package model

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// Setting represents a setting node in an ATV document.
type Setting struct {
	Pos               lexer.Position
	Name              string             `@Ident "="`
	SimpleValue       *SimpleValue       `( @@`
	TableValue        *TableValue        `| @@`
	ValueWithMetadata *ValueWithMetadata `| @@ )`
}

// Dupe returns a deep copy of the setting.
func (setting *Setting) Dupe() *Setting {

	if setting == nil {
		return nil
	}

	var copy = &Setting{Name: setting.Name}
	if setting.SimpleValue != nil {
		copy.SimpleValue = setting.SimpleValue.Dupe()
	} else if setting.ValueWithMetadata != nil {
		copy.ValueWithMetadata = setting.ValueWithMetadata.Dupe()
	} else if setting.TableValue != nil {
		copy.TableValue = setting.TableValue.Dupe()
	} else {
		panic("Unhandled setting value")
	}

	return copy
}

// ClearValue sets the value of all setting value types to nil (only one of them should be initialized).
func (setting *Setting) ClearValue() {

	if setting == nil {
		return
	}

	setting.SimpleValue = nil
	setting.ValueWithMetadata = nil
	setting.TableValue = nil
}

// String returns the setting as a string.
func (setting *Setting) String() string {

	if setting == nil {
		return "<nil>"
	}

	builder := strings.Builder{}
	err := setting.WriteDocumentPart(&builder, 0)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	return strings.TrimSpace(builder.String())
}

// MergeInto merges the current setting into the specified document.
func (setting *Setting) mergeInto(doc *Document, parentName string) error {

	if setting == nil {
		return nil
	}

	if setting.SimpleValue != nil || setting.ValueWithMetadata != nil {
		// top level simple value
		// => simply overwrite the setting in the document
		if err := doc.SetSetting(setting); err != nil {
			return err // document seems to contain a value with that name, but a different type (not a simple/complex value)
		}

	} else if setting.TableValue != nil {
		// a table value => merge table settings
		if err := doc.MergeTableSetting(setting); err != nil {
			return err // document seems to contain a value with that name, but a different type (not a table)
		}

	} else {
		panic("Unhandled setting type")
	}

	return nil
}

// GetRowReferences returns all row references recursively.
func (setting *Setting) GetRowReferences() []RowRef {

	if setting != nil {
		var allRowRefs []RowRef
		parts := []GetRowReferences{setting.SimpleValue, setting.ValueWithMetadata, setting.TableValue}
		for _, part := range parts {
			allRowRefs = append(allRowRefs, part.GetRowReferences()...)
		}
		return allRowRefs
	}

	return []RowRef{}
}

// GetRowIDs returns all row ids recursively.
func (setting *Setting) GetRowIDs() []RowID {

	if setting == nil {
		return []RowID{}
	}

	var allRowIDs []RowID
	parts := []GetRowIDs{setting.SimpleValue, setting.ValueWithMetadata, setting.TableValue}
	for _, part := range parts {
		allRowIDs = append(allRowIDs, part.GetRowIDs()...)
	}

	return allRowIDs
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (setting *Setting) WriteDocumentPart(writer *strings.Builder, indent int) error {

	// write the name of the setting
	line := fmt.Sprintf("%s%s = ", spacer(indent), setting.Name)
	_, err := writer.WriteString(line)
	if err != nil {
		return err
	}

	// write the setting value
	parts := []DocumentWriter{setting.SimpleValue, setting.ValueWithMetadata, setting.TableValue}
	for _, part := range parts {

		if !isNil(part) {
			err := part.WriteDocumentPart(writer, indent)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
