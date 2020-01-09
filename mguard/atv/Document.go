package atv

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	log "github.com/sirupsen/logrus"
)

// DocumentWriter is implemented by ATV document nodes that control how they are persisted.
type DocumentWriter interface {
	WriteDocumentPart(writer *strings.Builder, indent int) error
}

// GetRowReferences is implemented by ATV document setting nodes to return row references recursively.
type GetRowReferences interface {
	GetRowReferences() []*RowRef
}

// GetRowIDs is implemented by ATV document setting nodes to return row ids recursively.
type GetRowIDs interface {
	GetRowIDs() []RowID
}

// RowID is the id of a table row in an ATV diocument.
type RowID string

// Document represents a mGuard configuration document.
type Document struct {
	Root *DocumentRoot
}

// DocumentFromFile reads the specified ATV file from disk.
func DocumentFromFile(path string) (*Document, error) {

	// open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read the ATV file
	return DocumentFromReader(file)
}

// DocumentFromReader reads an ATV document from the specified io.Reader.
func DocumentFromReader(reader io.Reader) (*Document, error) {

	doc := &Document{}

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

func (doc *Document) parse(data string) error {

	// let the document always end with a new line to avoid handling EOF and EOL separately
	data += "\n"

	// build the parser
	root := &DocumentRoot{}
	parser, err := participle.Build(
		root,
		participle.Lexer(lexerDefinition),
		UnquoteToken("String"),
		participle.UseLookahead(2),
		participle.Elide("Whitespace", "Comment", "EOL"),
	)
	if err != nil {
		return err
	}

	// parse the document
	err = parser.Parse(strings.NewReader(data), root)
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
			repr.String(root, repr.Indent("  "), repr.OmitEmpty(true)))
	*/

	doc.Root = root
	return nil
}

// Dupe returns a copy of the ATV document.
func (doc *Document) Dupe() *Document {

	if doc == nil {
		return nil
	}

	buffer := bytes.Buffer{}
	err := doc.ToWriter(&buffer)
	if err != nil {
		// should not occur...
		panic("Unexpected error when serializing the ATV document")
	}

	other, err := DocumentFromReader(&buffer)
	if err != nil {
		// should not occur...
		panic("Unexpected error when deserializing the ATV document")
	}

	return other
}

// SetSetting sets the value of the setting with the specified name.
func (doc *Document) SetSetting(setting *Setting) error {

	copy := setting.Dupe()
	for _, node := range doc.Root.Nodes {
		if node.Setting != nil {
			if node.Setting.Name == copy.Name {
				if node.Setting.SimpleValue != nil && copy.SimpleValue != nil {

					x := node.Setting.SimpleValue.String()
					y := copy.SimpleValue.String()
					if x != y {
						log.Debugf("Simple value '%s' changed from '%s' to '%s'.", copy.Name, x, y)
						node.Setting.SimpleValue = copy.SimpleValue
					} else {
						log.Debugf("Simple value '%s' unchanged. Value is still '%s'.", copy.Name, x)
					}
					return nil

				} else if node.Setting.ComplexValue != nil && copy.ComplexValue != nil {

					x := node.Setting.ComplexValue.String()
					y := copy.ComplexValue.String()
					if x != y {
						log.Debugf("Complex value '%s' changed\n--from--\n%s\n--to--\n%s", copy.Name, x, y)
						node.Setting.ComplexValue = copy.ComplexValue
					} else {
						log.Debugf("Complex value '%s' unchanged. Value is still:\n%s", copy.Name, x)
					}
					return nil

				} else if node.Setting.RowRef != nil && copy.RowRef != nil {

					x := node.Setting.RowRef.String()
					y := copy.RowRef.String()
					if x != y {
						log.Debugf("RowRef value '%s' changed from '%s' to '%s'.", copy.Name, x, y)
						node.Setting.RowRef = copy.RowRef
					} else {
						log.Debugf("RowRef value '%s' unchanged. Value is still '%s'", copy.Name, x)
					}
					return nil

				} else if node.Setting.TableValue != nil && copy.TableValue != nil {

					x := node.Setting.TableValue.String()
					y := copy.TableValue.String()
					if x != y {
						log.Debugf("Table value '%s' changed\n--from--\n%s\n--to--\n%s", copy.Name, x, y)
						node.Setting.TableValue = copy.TableValue
					} else {
						log.Debugf("Table value '%s' unchanged. Value is still:\n%s", copy.Name, x)
					}
					return nil

				} else {

					return fmt.Errorf("Setting variable '%s' failed (type mismatch)", copy.Name)
				}
			}
		}
	}

	// specified setting was not found
	// => add it at the end
	newNode := &DocumentNode{Setting: copy}
	doc.Root.Nodes = append(doc.Root.Nodes, newNode)
	return nil
}

// MergeTableSetting replaces existing rows (same row id), appends rows that do not exist to the existing table
// or adds a new table value, if the table does not exist at all.
func (doc *Document) MergeTableSetting(setting *Setting) error {

	if setting.TableValue == nil {
		return fmt.Errorf("Specified setting '%s' is not a table value", setting.Name)
	}

	copy := setting.Dupe()
	for _, node := range doc.Root.Nodes {
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
			}
		}
	}

	// specified setting was not found
	// => add it at the end
	newNode := &DocumentNode{Setting: copy}
	doc.Root.Nodes = append(doc.Root.Nodes, newNode)
	return nil
}

// GetRowReferences returns all row references recursively.
func (doc *Document) GetRowReferences() []*RowRef {

	if doc == nil {
		return []*RowRef{}
	}

	var allRowRefs []*RowRef
	for _, node := range doc.Root.Nodes {
		allRowRefs = append(allRowRefs, node.Setting.GetRowReferences()...)
	}

	return allRowRefs
}

// GetRowIDs returns all row ids recursively.
func (doc *Document) GetRowIDs() []RowID {

	if doc == nil {
		return []RowID{}
	}

	var allRowIDs []RowID
	for _, node := range doc.Root.Nodes {
		allRowIDs = append(allRowIDs, node.Setting.GetRowIDs()...)
	}

	return allRowIDs
}

// ToFile saves the ATV document to the specified file.
func (doc *Document) ToFile(path string) error {

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
func (doc *Document) ToWriter(writer io.Writer) error {
	content := doc.String()
	_, err := writer.Write([]byte(content))
	return err
}

// String returns a properly formatted string representation of the ATV document.
func (doc *Document) String() string {

	if doc == nil {
		return "<nil>"
	}

	var builder strings.Builder
	doc.Root.WriteDocumentPart(&builder, 0)
	return builder.String()
}

// Merge merges the specified ATV document into the current one.
func (doc *Document) Merge(other *Document) (*Document, error) {

	for _, otherNode := range other.Root.Nodes {
		err := otherNode.Setting.mergeInto(doc, "")
		if err != nil {
			return nil, err
		}
	}
	return doc, nil
}

// UnquoteToken unquotes the specified token.
func UnquoteToken(types ...string) participle.Option {
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
