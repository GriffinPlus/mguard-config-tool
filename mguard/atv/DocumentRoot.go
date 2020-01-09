package atv

import (
	"reflect"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// DocumentRoot represents the root node of an ATV configuration document.
type DocumentRoot struct {
	Pos   lexer.Position
	Nodes []*DocumentNode `( @@ )*`
}

// Dupe returns a copy of the document root.
func (root *DocumentRoot) Dupe() *DocumentRoot {

	if root == nil {
		return nil
	}

	var nodesCopy []*DocumentNode
	for _, node := range root.Nodes {
		nodesCopy = append(nodesCopy, node.Dupe())
	}

	return &DocumentRoot{
		Nodes: nodesCopy,
	}
}

// GetRowReferences returns all row references recursively.
func (root *DocumentRoot) GetRowReferences() []*RowRef {

	if root == nil {
		return []*RowRef{}
	}

	var allRowRefs []*RowRef
	for _, node := range root.Nodes {
		allRowRefs = append(allRowRefs, node.GetRowReferences()...)
	}
	return allRowRefs
}

// GetRowIDs returns all row ids recursively.
func (root *DocumentRoot) GetRowIDs() []RowID {

	if root == nil {
		return []RowID{}
	}

	var allRowIDs []RowID
	for _, node := range root.Nodes {
		allRowIDs = append(allRowIDs, node.GetRowIDs()...)
	}

	return allRowIDs
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (root *DocumentRoot) WriteDocumentPart(writer *strings.Builder, indent int) error {

	var lastNodeType reflect.Type
	for _, node := range root.Nodes {

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
