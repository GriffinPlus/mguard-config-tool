package atv

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/alecthomas/participle"
)

// DocumentWriter is implemented by ATV document nodes that control how they are persisted.
type DocumentWriter interface {
	WriteDocumentPart(writer *strings.Builder, indent int) error
}

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
		participle.Unquote("String"),
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
	var builder strings.Builder
	doc.Root.WriteDocumentPart(&builder, 0)
	return builder.String()
}

// Merge merges the specified ATV document into the current one.
func (doc *Document) Merge(other *Document) (*Document, error) {
	return doc, nil
}
