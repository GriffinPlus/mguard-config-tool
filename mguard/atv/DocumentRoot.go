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
