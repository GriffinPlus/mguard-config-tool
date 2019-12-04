package atv

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"
	"github.com/alecthomas/repr"
	log "github.com/sirupsen/logrus"
)

var atvLexer = lexer.Must(ebnf.New(`
		Comment = "//" { "\u0000"…"\uffff"-"\n"-"\r" } .
		Pragma = "#" { alpha } { "\u0000"…"\uffff"-"\n"-"\r" } .
		String = "\"" { "\u0000"…"\uffff"-"\""-"\\" | "\\" any } "\"" .
		Ident = ( alpha ) { alpha | digit | "." | "_" } .
		EOL = ( "\n" | "\r" ) { "\n" | "\r" } .
		Assign = "=" .
		Whitespace = ( " " | "\t" ) { " " | "\t" } .
		CurlyBraceOpen = "{" .
		CurlyBraceClose = "}" .
		alpha = "a"…"z" | "A"…"Z" .
		digit = "0"…"9" .
		any = "\u0000"…"\uffff" .
		`))

// DocumentRoot represents the root node of an ATV configuration document.
type DocumentRoot struct {
	Pos   lexer.Position
	Nodes []*DocumentNode `( @@ )*`
}

// DocumentNode represents an element in an ATV configuration document.
type DocumentNode struct {
	Pos     lexer.Position
	Comment *Comment `( @@`
	Pragma  *Pragma  `| @Pragma` // needs extra conditioning step
	Setting *Setting `| @@)`
}

// Comment represents a comment in an ATV configuration document.
type Comment struct {
	Pos  lexer.Position
	Text string `@Comment`
}

// Pragma represents a pragma in an ATV configuration document.
type Pragma struct {
	Pos         lexer.Position
	PragmaName  string
	PragmaValue string
}

var pragmaSplitterRegex = regexp.MustCompile(`#(\w+)(\s+(.*))?$`)

// Capture initialized the pragma object when the parser encounters a pragma token in an ATV document.
func (pragma *Pragma) Capture(values []string) error {
	match := pragmaSplitterRegex.FindStringSubmatch(values[0])
	if len(match) == 0 {
		return fmt.Errorf("Splitting pragma failed")
	}
	pragma.PragmaName = match[1]
	pragma.PragmaValue = match[2]
	return nil
}

// Setting represents a setting node in an ATV document.
type Setting struct {
	Pos         lexer.Position
	Name        string  `@Ident "="`
	SimpleValue *string `( @String`
	TableValue  *Table  `| @@`
	RowRef      *RowRef `| @@ )`
}

// Table represents a table node in an ATV document.
type Table struct {
	Pos  lexer.Position
	UUID *string     `"{" ( "uuid" "=" @String )?`
	Rows []*TableRow `@@* "}"`
}

// TableRow represents a table row in an ATV document.
type TableRow struct {
	Pos   lexer.Position
	RowID *string    `"{" ( "{" "rid" "=" @String "}" )?`
	Items []*Setting `@@* "}"`
}

// RowRef represents a row reference in an ATV document.
type RowRef struct {
	Pos   lexer.Position
	RowID string `"{" "rowref" "=" @String "}"`
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
		participle.Lexer(atvLexer),
		participle.Unquote("String"),
		participle.UseLookahead(2),
		participle.Elide("Whitespace", "EOL"),
	)
	if err != nil {
		return err
	}

	// parse the document
	err = parser.Parse(strings.NewReader(data), root)
	if err != nil {
		return err
	}

	// print the document to the log
	log.Debugf(
		"Document Structure:"+
			"\n--------------------------------------------------------------------------------"+
			"\n%s"+
			"\n--------------------------------------------------------------------------------",
		repr.String(root, repr.Indent("  "), repr.OmitEmpty(true)))

	doc.Root = root
	return nil
}
