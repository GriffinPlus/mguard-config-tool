package atv

import (
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// documentNode represents an element in an ATV configuration document.
type documentNode struct {
	Pos     lexer.Position
	Pragma  *documentPragma  `( @Pragma` // needs extra conditioning step
	Setting *documentSetting `| @@)`
}

// Dupe returns a copy of the document node.
func (node *documentNode) Dupe() *documentNode {

	if node == nil {
		return nil
	}

	return &documentNode{
		Pragma:  node.Pragma.Dupe(),
		Setting: node.Setting.Dupe(),
	}
}

// GetRowReferences returns all row references recursively.
func (node *documentNode) GetRowReferences() []RowRef {

	if node != nil {
		return node.Setting.GetRowReferences()
	}

	return []RowRef{}
}

// GetRowIDs returns all row ids recursively.
func (node *documentNode) GetRowIDs() []RowID {

	if node == nil {
		return []RowID{}
	}

	return node.Setting.GetRowIDs()
}

// actual returns the actual node as DocumentNode itself is a node only encapsulating
// other nodes without being part of the document.
func (node *documentNode) actual() documentWriter {

	if node.Pragma != nil {
		return node.Pragma
	}

	if node.Setting != nil {
		return node.Setting
	}

	panic("Unhandled node type")
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (node *documentNode) WriteDocumentPart(writer *strings.Builder, indent int) error {

	children := []documentWriter{
		node.Pragma,
		node.Setting,
	}

	for _, child := range children {
		if !isNil(child) {
			err := child.WriteDocumentPart(writer, indent)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
