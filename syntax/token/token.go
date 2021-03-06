package token

// A lot of this package was shamefully stolen from
// https://cs.opensource.google/go/go/+/refs/tags/go1.17:src/go/token/token.go
// It should be rewritten at some point.

import (
	"github.com/gala377/MLLang/syntax/span"
)

type Id = int

type Token struct {
	Typ  Id
	Val  string
	Span *span.Span
}

const (
	Error Id = iota
	Identifier
	InfixIdentifier
	InfixNotModuledIdentifier
	Integer
	Float
	String
	Colon
	Comma
	LParen
	RParen
	LBracket
	RBracket
	LSquareParen
	RSquareParen
	Indent
	NewLine
	Operator
	Comment
	Pipe
	Access
	Quote

	keywords_beg
	Fn
	If
	Else
	While
	Let
	True
	False
	Do
	None
	Return
	Handle
	With
	Effect
	Resume
	keywords_end

	operators_beg
	Assignment
	Exclamation
	Arrow
	Dollar
	operators_end

	Eof
)

func IdToString(id Id) string {
	return tokens[id]
}

var tokens = [...]string{
	Error:                     "ERROR",
	Identifier:                "IDENT",
	InfixIdentifier:           "INFIX_IDENT",
	InfixNotModuledIdentifier: "INFIX_NOTMODULED_IDENT",
	Integer:                   "INT",
	Float:                     "FLOAT",
	String:                    "STRING",
	Comment:                   "COMMENT",
	Colon:                     ":",
	Comma:                     ",",
	LParen:                    "(",
	RParen:                    ")",
	LBracket:                  "{",
	RBracket:                  "}",
	LSquareParen:              "[",
	RSquareParen:              "]",
	Pipe:                      "|",
	Indent:                    "INDENT",
	NewLine:                   "NEWLINE",
	Operator:                  "OPERATOR",
	Access:                    ".",
	Quote:                     "`",

	Fn:     "fn",
	If:     "if",
	Else:   "else",
	While:  "while",
	Let:    "let",
	True:   "true",
	False:  "false",
	None:   "none",
	Return: "return",
	Handle: "handle",
	With:   "with",
	Effect: "effect",
	Do:     "do",
	Resume: "resume",

	Assignment:  "=",
	Exclamation: "!",
	Arrow:       "->",
	Dollar:      "$",

	Eof: "EOF",
}

var keywords map[string]Id

var operators map[string]Id

func init() {
	keywords = make(map[string]Id)
	for i := keywords_beg + 1; i < keywords_end; i++ {
		keywords[tokens[i]] = i
	}
	operators = make(map[string]Id)
	for i := operators_beg + 1; i < operators_end; i++ {
		operators[tokens[i]] = i
	}
}

func IsKeyword(name string) bool {
	_, ok := keywords[name]
	return ok
}

func Lookup(ident string) Id {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	return Identifier
}

func LookupOperator(op string) Id {
	if tok, isop := operators[op]; isop {
		return tok
	}
	return Operator
}

func NewEof(pos span.Position) Token {
	span := span.NewSpan(pos, pos)
	return Token{
		Typ:  Eof,
		Val:  "",
		Span: &span,
	}
}
