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

	keywords_beg
	Fn
	Val
	If
	Else
	While
	Match
	Let
	Macro
	True
	False
	Do
	None
	keywords_end

	operators_beg
	Assignment
	Exclamation
	operators_end

	Eof
)

func IdToString(id Id) string {
	return tokens[id]
}

var tokens = [...]string{
	Error:        "ERROR",
	Identifier:   "IDENT",
	Integer:      "INT",
	Float:        "FLOAT",
	String:       "STRING",
	Comment:      "COMMENT",
	Colon:        ":",
	Comma:        ",",
	LParen:       "(",
	RParen:       ")",
	LBracket:     "{",
	RBracket:     "}",
	LSquareParen: "[",
	RSquareParen: "]",
	Pipe:         "|",
	Indent:       "INDENT",
	NewLine:      "NEWLINE",
	Operator:     "OPERATOR",
	Do:           "do",

	Fn:    "fn",
	Val:   "val",
	If:    "if",
	Else:  "else",
	While: "while",
	Match: "match",
	Let:   "let",
	Macro: "macro",
	True:  "true",
	False: "false",
	None:  "none",

	Assignment:  "=",
	Exclamation: "!",

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

func IsArrow(t *Token) bool {
	return t.Val == "->"
}
