package atv

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
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

// GetPragma gets the pragma with the specified name.
func (root *DocumentRoot) GetPragma(name string) *Pragma {
	for _, node := range root.Nodes {
		if node.Pragma != nil && node.Pragma.Name == name {
			return node.Pragma
		}
	}

	// pragma with the specified name does not exist
	return nil
}

// GetVersion gets the version of the document.
func (root *DocumentRoot) GetVersion() (*Version, error) {

	versionPragma := root.GetPragma("version")
	if versionPragma == nil {
		return nil, fmt.Errorf("The ATV document does not contain a version pragma")
	}

	versionRegex := regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)\.(.+)$`)
	matches := versionRegex.FindAllStringSubmatch(versionPragma.Value, -1)
	if matches == nil {
		return nil, fmt.Errorf("The ATV document does not contain a properly formatted version number")
	}

	major, _ := strconv.Atoi(matches[0][1])
	minor, _ := strconv.Atoi(matches[0][2])
	patch, _ := strconv.Atoi(matches[0][3])
	version := Version{Major: major, Minor: minor, Patch: patch}
	if len(matches[0]) > 3 {
		version.Suffix = matches[0][4]
	}

	return &version, nil
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
