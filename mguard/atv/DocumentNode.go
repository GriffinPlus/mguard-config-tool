package atv

import (
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// DocumentNode represents an element in an ATV configuration document.
type DocumentNode struct {
	Pos     lexer.Position
	Pragma  *Pragma  `( @Pragma` // needs extra conditioning step
	Setting *Setting `| @@)`
}

// actual returns the actual node as DocumentNode itself is a node only encapsulating
// other nodes without being part of the document.
func (node *DocumentNode) actual() DocumentWriter {
	if node.Pragma != nil {
		return node.Pragma
	}
	if node.Setting != nil {
		return node.Setting
	}
	panic("Unhandled node type")
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (node *DocumentNode) WriteDocumentPart(writer *strings.Builder, indent int) error {

	children := []DocumentWriter{
		node.Pragma,
		node.Setting,
	}

	for _, child := range children {
		err := child.WriteDocumentPart(writer, indent)
		if err != nil {
			return err
		}
	}

	return nil
}
