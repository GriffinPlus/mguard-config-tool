package model

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// Setting represents a setting node in an ATV document.
type Setting struct {
	Pos          lexer.Position
	Name         string        `@Ident "="`
	SimpleValue  *SimpleValue  `( @@`
	TableValue   *TableValue   `| @@`
	RowRef       *RowRef       `| @@`
	ComplexValue *ComplexValue `| @@ )`
}

// Dupe returns a deep copy of the setting.
func (setting *Setting) Dupe() *Setting {

	if setting == nil {
		return nil
	}

	var copy = &Setting{Name: setting.Name}
	if setting.SimpleValue != nil {
		copy.SimpleValue = setting.SimpleValue.Dupe()
	} else if setting.ComplexValue != nil {
		copy.ComplexValue = setting.ComplexValue.Dupe()
	} else if setting.TableValue != nil {
		copy.TableValue = setting.TableValue.Dupe()
	} else if setting.RowRef != nil {
		copy.RowRef = setting.RowRef.Dupe()
	} else {
		panic("Unhandled setting value")
	}

	return copy
}

// MergeInto merges the current setting into the specified document.
func (setting *Setting) mergeInto(doc *Document, parentName string) error {

	if setting == nil {
		return nil
	}

	if setting.SimpleValue != nil || setting.ComplexValue != nil {
		// top level simple/complex value
		// => simply overwrite the setting in the document
		if err := doc.SetSetting(setting); err != nil {
			return err // document seems to contain a value with that name, but a different type (not a simple/complex value)
		}

	} else if setting.RowRef != nil {
		// a reference to a table row
		// => set reference (it should not be in the destination document)
		if err := doc.SetSetting(setting); err != nil {
			return err // document seems to contain a value with that name, but a different type (not a rowref)
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
func (setting *Setting) GetRowReferences() []*RowRef {

	if setting == nil {
		return nil
	}

	var allRowRefs []*RowRef
	parts := []GetRowReferences{setting.SimpleValue, setting.ComplexValue, setting.RowRef, setting.TableValue}
	for _, part := range parts {
		allRowRefs = append(allRowRefs, part.GetRowReferences()...)
	}

	return allRowRefs
}

// GetRowIDs returns all row ids recursively.
func (setting *Setting) GetRowIDs() []RowID {

	if setting == nil {
		return []RowID{}
	}

	var allRowIDs []RowID
	parts := []GetRowIDs{setting.SimpleValue, setting.ComplexValue, setting.RowRef, setting.TableValue}
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
	parts := []DocumentWriter{setting.SimpleValue, setting.ComplexValue, setting.RowRef, setting.TableValue}
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
