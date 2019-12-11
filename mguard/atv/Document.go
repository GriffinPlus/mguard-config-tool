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

// charsToQuote contains characters that are quoted in string in ATV documents.
const charsToQuote = `"\`

// quote quotes the specified string in the ATV specific fashion.
func quote(s string) string {
	quoted := strings.Builder{}
	quoted.WriteRune('"')
	for _, c := range s {
		if strings.ContainsRune(charsToQuote, c) {
			quoted.WriteRune('\\')
		}
		quoted.WriteRune(c)
	}
	quoted.WriteRune('"')
	return quoted.String()
}

// unquote unquotes the specified string in the ATV specific fashion.
func unquote(s string) (string, error) {

	// check whether the string contains the expected embracing quotes
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", fmt.Errorf("String is not properly quoted")
	}

	// strip embracing quotes
	s = s[1 : len(s)-1]

	unquoted := strings.Builder{}
	unquotePending := false
	for _, c := range s {
		if !unquotePending && c == '\\' {
			unquotePending = true
			continue
		}

		if unquotePending && !strings.ContainsRune(charsToQuote, c) {
			unquoted.WriteRune('\\')
		}

		unquoted.WriteRune(c)
		unquotePending = false
	}

	if unquotePending {
		return "", fmt.Errorf("Unpaired backslash found when unquoting string (%s)", s)
	}

	return unquoted.String(), nil
}
