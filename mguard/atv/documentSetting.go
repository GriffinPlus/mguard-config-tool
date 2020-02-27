package atv

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// documentSetting represents a setting node in an ATV document.
type documentSetting struct {
	Pos               lexer.Position
	Name              string                     `@Ident "="`
	SimpleValue       *documentSimpleValue       `( @@`
	TableValue        *documentTableValue        `| @@`
	ValueWithMetadata *documentValueWithMetadata `| @@ )`
}

// Dupe returns a deep copy of the setting.
func (setting *documentSetting) Dupe() *documentSetting {

	if setting == nil {
		return nil
	}

	var copy = &documentSetting{Name: setting.Name}
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
func (setting *documentSetting) ClearValue() {

	if setting == nil {
		return
	}

	setting.SimpleValue = nil
	setting.ValueWithMetadata = nil
	setting.TableValue = nil
}

// String returns the setting as a string.
func (setting *documentSetting) String() string {

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

// mergeInto merges the current setting into the specified document.
func (setting *documentSetting) mergeInto(doc *document) error {

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
func (setting *documentSetting) GetRowReferences() []RowRef {

	if setting != nil {
		var allRowRefs []RowRef
		parts := []getRowReferences{setting.SimpleValue, setting.ValueWithMetadata, setting.TableValue}
		for _, part := range parts {
			allRowRefs = append(allRowRefs, part.GetRowReferences()...)
		}
		return allRowRefs
	}
	return []RowRef{}
}

// GetRowIDs returns all row ids recursively.
func (setting *documentSetting) GetRowIDs() []RowID {

	if setting != nil {
		var allRowIDs []RowID
		parts := []getRowIDs{setting.SimpleValue, setting.ValueWithMetadata, setting.TableValue}
		for _, part := range parts {
			allRowIDs = append(allRowIDs, part.GetRowIDs()...)
		}
		return allRowIDs
	}
	return []RowID{}
}

// getSetting gets the setting at the specified path.
// If the setting does not exist, nil is returned.
func (setting *documentSetting) getSetting(path documentSettingPath, index int) (*documentSetting, error) {

	// abort, if the setting is found
	if index == len(path) {
		return setting, nil
	}

	if setting.SimpleValue != nil || setting.ValueWithMetadata != nil {
		return nil, fmt.Errorf("Setting '%s' is a single value, but the path '%s' specifies a more nested setting", path[0:index], path)
	}

	if setting.TableValue != nil {

		if path[index].row == nil {
			return nil, fmt.Errorf("Setting '%s' is a table value, but the path '%s' does not address a specific row", path[0:index], path)
		}

		if index+1 == len(path) {
			return nil, fmt.Errorf("Setting '%s' is a table value, but the path '%s' does not address a value within a row", path[0:index], path)
		}

		rowIndex := *path[index].row
		if rowIndex >= len(setting.TableValue.Rows) {
			return nil, nil
		}

		row := setting.TableValue.Rows[rowIndex]
		for _, item := range row.Items {
			if item.Name == *path[index+1].name {
				return item.getSetting(path, index+2)
			}
		}

		return nil, nil
	}

	panic("Unhandled setting type")
}

// GetValue gets the value of the setting, if the setting is a simple value - with or without metadata.
// If the setting has a table value, an error is returned.
func (setting *documentSetting) GetValue() (string, error) {

	if setting.SimpleValue != nil {
		return setting.SimpleValue.Value, nil
	}

	if setting.ValueWithMetadata != nil {
		var value string
		if setting.ValueWithMetadata.Data.TryGet("value", &value) {
			return value, nil
		}
		return "", fmt.Errorf("The setting does not contain a value")
	}

	return "", fmt.Errorf("The setting is not a single value")
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (setting *documentSetting) WriteDocumentPart(writer *strings.Builder, indent int) error {

	// write the name of the setting
	line := fmt.Sprintf("%s%s = ", spacer(indent), setting.Name)
	_, err := writer.WriteString(line)
	if err != nil {
		return err
	}

	// write the setting value
	parts := []documentWriter{setting.SimpleValue, setting.ValueWithMetadata, setting.TableValue}
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
