package atv

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	log "github.com/sirupsen/logrus"
)

// DocumentWriter is implemented by ATV document nodes that control how they are persisted.
type documentWriter interface {
	WriteDocumentPart(writer *strings.Builder, indent int) error
}

// GetRowReferences is implemented by ATV document setting nodes to return row references recursively.
type getRowReferences interface {
	GetRowReferences() []RowRef
}

// GetRowIDs is implemented by ATV document setting nodes to return row ids recursively.
type getRowIDs interface {
	GetRowIDs() []RowID
}

// RowID is the id of a table row in an ATV document.
type RowID string

// RowRef represents a reference to a table row.
type RowRef string

// UUID represents a UUID associated with a setting.
type UUID string

// document represents a mGuard configuration document.
type document struct {
	Pos   lexer.Position
	Nodes []*documentNode `( @@ )*`
}

// FromFile reads the specified ATV file from disk.
func documentFromFile(path string) (*document, error) {

	// open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read the ATV file
	return documentFromReader(file)
}

// FromReader reads an ATV document from the specified io.Reader.
func documentFromReader(reader io.Reader) (*document, error) {

	doc := &document{}

	// read file into buffer
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// parse the document
	docData := strings.ReplaceAll(string(buf), "\r\n", "\n")
	err = doc.parse(docData)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (doc *document) parse(data string) error {

	// let the document always end with a new line to avoid handling EOF and EOL separately
	data += "\n"

	// build the parser
	newDoc := &document{}
	parser, err := participle.Build(
		newDoc,
		participle.Lexer(lexerDefinition),
		unquoteToken("String"),
		participle.UseLookahead(2),
		participle.Elide("Whitespace", "Comment", "EOL"),
	)
	if err != nil {
		return err
	}

	// parse the document
	err = parser.Parse(strings.NewReader(data), newDoc)
	if err != nil {
		return err
	}

	/*
		// print the document to the log
		log.Debugf(
			"Document Structure:"+
				"\n--------------------------------------------------------------------------------"+
				"\n%s"+
				"\n--------------------------------------------------------------------------------",
			repr.String(*newDoc, repr.Indent("  "), repr.OmitEmpty(true)))
	*/

	*doc = *newDoc
	return nil
}

// Dupe returns a copy of the ATV document.
func (doc *document) Dupe() *document {

	if doc == nil {
		return nil
	}

	var nodesCopy []*documentNode
	for _, node := range doc.Nodes {
		nodesCopy = append(nodesCopy, node.Dupe())
	}

	return &document{
		Nodes: nodesCopy,
	}
}

// GetPragma gets the pragma with the specified name.
func (doc *document) GetPragma(name string) (*documentPragma, error) {

	for _, node := range doc.Nodes {
		if node.Pragma != nil && node.Pragma.Name == name {
			return node.Pragma, nil
		}
	}

	// pragma with the specified name does not exist
	return nil, nil
}

// SetPragma sets the pragma with the specified name.
func (doc *document) SetPragma(name string, value string) (*documentPragma, error) {

	for _, node := range doc.Nodes {
		if node.Pragma != nil && node.Pragma.Name == name {
			node.Pragma.Value = value
			return node.Pragma, nil
		}
	}

	// pragma with the specified name does not exist
	// => add it after the last pragma (there is always at least a version pragma)
	lastPragmaIndex := -1
	for i, node := range doc.Nodes {
		if node.Pragma != nil {
			lastPragmaIndex = i
		}
	}
	doc.Nodes = append(doc.Nodes, nil)
	copy(doc.Nodes[lastPragmaIndex+2:], doc.Nodes[lastPragmaIndex+1:])
	pragma := &documentPragma{Name: name, Value: value}
	doc.Nodes[lastPragmaIndex+1] = &documentNode{Pragma: pragma}
	return pragma, nil
}

// GetUUID gets the uuid associated with the specified setting.
func (doc *document) GetUUID(settingName string) (*UUID, error) {

	value, err := doc.GetAttribute(settingName, "uuid")
	if err != nil || value == nil {
		return nil, err
	}

	uuid := UUID(*value)
	return &uuid, nil
}

// SetUUID sets the uuid associated with the specified setting.
func (doc *document) SetUUID(settingName string, uuid UUID) error {
	return doc.SetAttribute(settingName, "uuid", string(uuid))
}

// RemoveUUID removes the UUID of the setting with the specified name.
func (doc *document) RemoveUUID(settingName string) error {
	return doc.RemoveAttribute(settingName, "uuid")
}

// GetAccess gets the access modifier of the setting with the specified name.
func (doc *document) GetAccess(settingName string) (*AccessModifier, error) {

	value, err := doc.GetAttribute(settingName, "uuid")
	if err != nil || value == nil {
		return nil, err
	}

	access, err := ParseAccessModifier(*value)
	if err != nil {
		return nil, err
	}

	return &access, nil
}

// SetAccess sets the access modifier of the setting with the specified name.
func (doc *document) SetAccess(settingName string, access AccessModifier) error {
	return doc.SetAttribute(settingName, "access", string(access))
}

// RemoveAccess removes the access modifier from the setting with the specified name.
func (doc *document) RemoveAccess(settingName string) error {
	return doc.RemoveAttribute(settingName, "access")
}

// GetAttribute gets the specified attribute associated with the specified setting.
func (doc *document) GetAttribute(settingName, attributeName string) (*string, error) {

	setting, err := doc.GetSetting(settingName)
	if err != nil {
		return nil, err
	}

	if setting != nil {

		var dict *dictionary
		if setting.ValueWithMetadata != nil {
			dict = &setting.ValueWithMetadata.Data
		} else if setting.TableValue != nil {
			dict = &setting.TableValue.Attributes
		}

		if dict != nil {
			var value string
			if dict.TryGet(attributeName, &value) {
				return &value, nil
			}

			return nil, nil
		}

		panic("Unhandled value type")
	}

	return nil, nil
}

// SetAttribute sets the specified attribute associated with the specified setting.
func (doc *document) SetAttribute(settingName, attributeName, attributeValue string) error {

	setting, err := doc.GetSetting(settingName)
	if err != nil {
		return err
	}

	if setting != nil {
		if setting.SimpleValue != nil {

			items := dictionary{
				keyValuePair{attributeName, string(attributeValue)},
				keyValuePair{"value", setting.SimpleValue.Value},
			}
			setting.ValueWithMetadata = &documentValueWithMetadata{Data: items}
			setting.SimpleValue = nil
			return nil
		}

		var dict *dictionary
		if setting.ValueWithMetadata != nil {
			dict = &setting.ValueWithMetadata.Data
		} else if setting.TableValue != nil {
			dict = &setting.TableValue.Attributes
		}

		if dict != nil {
			dict.Set(attributeName, attributeValue)
			return nil
		}

		panic("Unhandled value type")
	}

	return fmt.Errorf("Setting '%s' does not exist", settingName)
}

// RemoveAttribute removes an attribute from a setting with the specified name.
func (doc *document) RemoveAttribute(settingName, attributeName string) error {

	setting, err := doc.GetSetting(settingName)
	if err != nil {
		return err
	}

	if setting != nil {

		if setting.SimpleValue != nil {
			return nil
		}

		if setting.ValueWithMetadata != nil {

			dict := &setting.ValueWithMetadata.Data
			if dict.Remove(attributeName) {
				if len(*dict) == 1 && (*dict)[0].Key == "value" {
					setting.SimpleValue = &documentSimpleValue{Value: (*dict)[0].Value}
					setting.ValueWithMetadata = nil
				}
			}
			return nil

		} else if setting.TableValue != nil {

			setting.TableValue.Attributes.Remove(attributeName)
			return nil

		}

		panic("Unhandled value type")
	}

	return fmt.Errorf("Setting '%s' does not exist", settingName)
}

// GetRowReferences returns all row references recursively.
func (doc *document) GetRowReferences() []RowRef {
	var allRowRefs []RowRef
	for _, node := range doc.Nodes {
		allRowRefs = append(allRowRefs, node.GetRowReferences()...)
	}
	return allRowRefs
}

// GetRowIDs returns all row ids recursively.
func (doc *document) GetRowIDs() []RowID {
	var allRowIDs []RowID
	for _, node := range doc.Nodes {
		allRowIDs = append(allRowIDs, node.GetRowIDs()...)
	}
	return allRowIDs
}

// ToFile saves the ATV document to the specified file.
func (doc *document) ToFile(path string) error {

	if doc == nil {
		return ErrNilReceiver
	}

	// create directories on the way, if necessary
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	// open the file for writing
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// write the ATV document
	return doc.ToWriter(file)
}

// ToWriter writes the ATV document to the specified io.Writer.
func (doc *document) ToWriter(writer io.Writer) error {

	if doc == nil {
		return ErrNilReceiver
	}

	content := doc.String()
	_, err := writer.Write([]byte(content))
	return err
}

// String returns a properly formatted string representation of the ATV document.
func (doc *document) String() string {

	if doc == nil {
		return "<nil>"
	}

	var builder strings.Builder
	doc.WriteDocumentPart(&builder, 0)
	return builder.String()
}

// SetSetting sets the value of the setting with the specified name.
func (doc *document) SetSetting(setting *documentSetting) error {

	copy := setting.Dupe()
	for _, node := range doc.Nodes {
		if node.Setting != nil {
			if node.Setting.Name == copy.Name {
				x := node.Setting.String()
				y := copy.String()
				if x != y {
					log.Debugf("Setting '%s' changed.\nFrom: %s\nTo:   %s", copy.Name, x, y)
					node.Setting.ClearValue()
					if copy.SimpleValue != nil {
						node.Setting.SimpleValue = copy.SimpleValue
					} else if copy.ValueWithMetadata != nil {
						node.Setting.ValueWithMetadata = copy.ValueWithMetadata
					} else if copy.TableValue != nil {
						node.Setting.TableValue = copy.TableValue
					} else {
						panic("Unhandled value type.")
					}
				} else {
					log.Debugf("Setting '%s' unchanged.\nValue: %s", copy.Name, x)
				}

				return nil
			}
		}
	}

	// specified setting was not found
	// => add it at the end
	newNode := &documentNode{Setting: copy}
	doc.Nodes = append(doc.Nodes, newNode)
	return nil
}

// MergeTableSetting replaces existing rows (same row id), appends rows that do not exist to the existing table
// or adds a new table value, if the table does not exist at all.
func (doc *document) MergeTableSetting(setting *documentSetting) error {

	if setting.TableValue == nil {
		return fmt.Errorf("Specified setting '%s' is not a table value", setting.Name)
	}

	copy := setting.Dupe()
	for _, node := range doc.Nodes {
		if node.Setting != nil {
			if node.Setting.Name == copy.Name {

				// found setting with the specified name
				// => abort, if the setting is not a table value
				if node.Setting.TableValue == nil {
					return fmt.Errorf("Setting '%s' in the document is not a table value", copy.Name)
				}

				// update table
			update_loop:
				for _, rowToSet := range copy.TableValue.Rows {

					// update existing row, if possible
					for i, existingRow := range node.Setting.TableValue.Rows {
						if existingRow.HasSameID(rowToSet) {
							x := existingRow.String()
							y := rowToSet.String()
							log.Debugf("Table value '%s' contains row with id '%s'. Row changed\n--from--\n%s\n--to--\n%s", setting.Name, *rowToSet.RowID, x, y)
							node.Setting.TableValue.Rows[i] = rowToSet
							continue update_loop
						}
					}

					// insert row, if there is no row with the same id, yet
					log.Debugf("Table value '%s' does not contain row to set, yet. Appending\n%s", setting.Name, rowToSet.String())
					node.Setting.TableValue.Rows = append(node.Setting.TableValue.Rows, rowToSet)
				}

				return nil
			}
		}
	}

	// specified setting was not found
	// => add it at the end
	newNode := &documentNode{Setting: copy}
	doc.Nodes = append(doc.Nodes, newNode)
	return nil
}

// Merge merges all settings of the specified ATV document into the current one.
func (doc *document) Merge(other *document) (*document, error) {
	return doc.MergeSelectively(other, nil)
}

// Merge merges the configured settings of the specified ATV document into the current one.
// config : The merge configuration (nil merges all settings)
func (doc *document) MergeSelectively(other *document, config *MergeConfiguration) (*document, error) {

	if doc == nil {
		return nil, ErrNilReceiver
	}

	copy := doc.Dupe()
	for _, otherNode := range other.Nodes {
		if otherNode.Setting != nil {
			otherSettingPath, _ := parseDocumentSettingPath(otherNode.Setting.Name) // works for top-level setting only!
			if config == nil || config.ShouldMergeSetting(otherSettingPath) {
				log.Infof("Merging setting '%s'...", otherNode.Setting.Name)
				err := otherNode.Setting.mergeInto(copy)
				if err != nil {
					return nil, err
				}
			} else {
				log.Debugf("Setting '%s' is not in merge list. Skipping...", otherNode.Setting.Name)
			}
		}
	}

	return copy, nil
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (doc *document) WriteDocumentPart(writer *strings.Builder, indent int) error {

	var lastNodeType reflect.Type
	for _, node := range doc.Nodes {

		// insert an extra newline to separate nodes of different types
		nodeType := reflect.TypeOf(node.actual())
		if nodeType != lastNodeType {
			writer.WriteString("\n")
		}
		lastNodeType = nodeType

		// write node
		err := node.WriteDocumentPart(writer, indent)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetSetting finds the setting with the specified name.
// If the setting does not exist, nil is returned.
func (doc *document) GetSetting(settingName string) (*documentSetting, error) {

	path, err := parseDocumentSettingPath(settingName)
	if err != nil {
		return nil, err
	}

	setting, err := doc.getSetting(path)
	if err != nil {
		return nil, err
	}

	return setting, nil
}

// getSetting gets the setting at the specified path.
// If the setting does not exist, nil is returned.
func (doc *document) getSetting(path documentSettingPath) (*documentSetting, error) {

	if len(path) == 0 {
		return nil, fmt.Errorf("Path is empty")
	}

	if path[0].name == nil {
		return nil, fmt.Errorf("The first path token '%s' is not a setting name", path[0])
	}

	for _, node := range doc.Nodes {
		if node.Setting != nil {
			if node.Setting.Name == *path[0].name {
				return node.Setting.getSetting(path, 1)
			}
		}
	}

	// specified setting does not exist
	return nil, nil
}

// SetSimpleValueSetting sets the setting with the specified name to a simple string value.
func (doc *document) SetSimpleValueSetting(settingName string, value string) error {

	path, err := parseDocumentSettingPath(settingName)
	if err != nil {
		return err
	}

	err = doc.setSimpleValueSetting(path, value)
	if err != nil {
		return err
	}

	return nil
}

// setSimpleValueSetting sets the setting at the specified path to the specified simple string value
func (doc *document) setSimpleValueSetting(path documentSettingPath, value string) error {

	if len(path) == 0 {
		return fmt.Errorf("Path is empty")
	}

	if path[0].name == nil {
		return fmt.Errorf("The first path token '%s' is not a setting name", path[0])
	}

	// try to get existing setting
	setting, err := doc.getSetting(path)
	if err != nil {
		return err
	}

	// create the path to the setting, if the setting does not exist, yet
	if setting == nil {
		setting, err = doc.createSettingPlaceholder(path)
		if err != nil {
			return err
		}
	}

	// set setting value
	setting.ClearValue()
	setting.SimpleValue = &documentSimpleValue{Value: value}
	return nil
}

// RemoveSetting removes the setting with the specified name.
// If the setting does not exist, no error is signalled.
func (doc *document) RemoveSetting(settingName string) error {

	path, err := parseDocumentSettingPath(settingName)
	if err != nil {
		return err
	}

	return doc.removeSetting(path)
}

// removeSetting removes the setting at the specified path.
// If the setting does not exist, no error is signalled.
func (doc *document) removeSetting(path documentSettingPath) error {

	if len(path) == 0 {
		return fmt.Errorf("Path is empty")
	}

	if path[0].name == nil {
		return fmt.Errorf("The first path token '%s' is not a setting name", path[0])
	}

	for i, node := range doc.Nodes {
		if node.Setting != nil {
			if node.Setting.Name == *path[0].name {
				if len(path) == 1 {
					// top-level item => just remove it
					doc.Nodes = append(doc.Nodes[:i], doc.Nodes[i+1:]...)
					return nil
				} else {
					// nested setting => dive deeper
					return node.Setting.removeSetting(path, 1)
				}
			}
		}
	}

	// specified setting does not exist
	return nil
}

// createSettingPlaceholder creates all necessary nodes in the document to the setting at the specified path.
// The setting itself is empty, if it did not exist before.
func (doc *document) createSettingPlaceholder(path documentSettingPath) (*documentSetting, error) {

	if len(path) == 0 {
		return nil, fmt.Errorf("Path is empty")
	}

	if path[0].name == nil {
		return nil, fmt.Errorf("The first path token '%s' is not a setting name", path[0])
	}

	// use existing node, if it exists already
	for _, node := range doc.Nodes {
		if node.Setting != nil {
			if node.Setting.Name == *path[0].name {
				if len(path) == 1 {
					// a top-level setting
					// => that's already what we're looking for
					return node.Setting, nil
				} else {
					// a nested setting => dive deeper
					return node.Setting.createSettingPlaceholder(path, 1)
				}
			}
		}
	}

	// the setting was not set, because there was no existing setting with the specified name
	// => add a new one
	if len(path) == 1 {
		// a top-level setting
		newSetting := &documentSetting{Name: *path[0].name}
		newNode := &documentNode{Setting: newSetting}
		doc.Nodes = append(doc.Nodes, newNode)
		return newSetting, nil
	} else {
		// a nested setting
		newSetting := &documentSetting{
			Name: *path[0].name,
			TableValue: &documentTableValue{
				Rows: []*documentTableRow{}}}
		newNode := &documentNode{Setting: newSetting}
		doc.Nodes = append(doc.Nodes, newNode)
		return newNode.Setting.createSettingPlaceholder(path, 1)
	}
}

// UnquoteToken unquotes the specified token.
func unquoteToken(types ...string) participle.Option {
	if len(types) == 0 {
		types = []string{"String"}
	}
	return participle.Map(func(t lexer.Token) (lexer.Token, error) {
		value, err := unquote(t.Value)
		if err != nil {
			return t, lexer.Errorf(t.Pos, "invalid quoted string %q: %s", t.Value, err.Error())
		}
		t.Value = value
		return t, nil
	}, types...)
}
