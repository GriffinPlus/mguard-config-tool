package atv

import (
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/ebnf"
)

// lexerDefinition is an EBNF based participle lexer definition for ATV documents.
var lexerDefinition = lexer.Must(ebnf.New(`
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
