package atv

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/participle/lexer"
)

// Comment represents a comment in an ATV configuration document.
type Comment struct {
	Pos  lexer.Position
	Text string
}

var commentSplitterRegex = regexp.MustCompile(`//(.*)?`)

// Capture initializes the comment object when the parser encounters a comment token in an ATV document.
func (comment *Comment) Capture(values []string) error {
	match := commentSplitterRegex.FindStringSubmatch(values[0])
	if len(match) == 0 {
		return fmt.Errorf("Splitting comment failed")
	}
	comment.Text = match[1]
	return nil
}

// WriteDocumentPart writes a part of the ATV document to the specified writer.
func (comment *Comment) WriteDocumentPart(writer *strings.Builder, indent int) error {
	if comment == nil {
		return nil
	}

	line := fmt.Sprintf("%s//%s\n", spacer(indent), comment.Text)
	_, err := writer.WriteString(line)
	return err
}
