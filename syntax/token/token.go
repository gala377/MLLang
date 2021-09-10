package token

// A lot of this package was shamefully stolen from
// https://cs.opensource.google/go/go/+/refs/tags/go1.17:src/go/token/token.go
// It should be rewritten at some point.

import (
	"unicode"

	"github.com/gala377/MLLang/syntax/span"
)

type Id = int

type Token struct {
	Typ  Id
	Val  string
	Span span.Span
}

const (
	Unknown Id = iota
	Identifier
	Integer
	Float
	String
	Colon
	Assignment
	LParen
	RParen
	LBracket
	RBracket
	LSquareParen
	RSquareParen
	Indent
	NewLine

	keywords_beg
	Fn
	Val
	If
	Else
	While
	Match
	Let
	Macro
	keywords_end

	Eof
)

var tokens = [...]string{
	Unknown:      "UNKNOWN",
	Identifier:   "IDENT",
	Integer:      "INT",
	Float:        "FLOAT",
	String:       "STRING",
	Colon:        ":",
	Assignment:   "=",
	LParen:       "(",
	RParen:       ")",
	LBracket:     "{",
	RBracket:     "}",
	LSquareParen: "[",
	RSquareParen: "]",
	Indent:       "INDENT",
	NewLine:      "NEWLINE",

	Fn:    "fn",
	Val:   "val",
	If:    "if",
	Else:  "else",
	While: "while",
	Match: "match",
	Let:   "let",
	Macro: "macro",
}

var keywords map[string]Id

func init() {
	keywords = make(map[string]Id)
	for i := keywords_beg + 1; i < keywords_end; i++ {
		keywords[tokens[i]] = i
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

func IsValidIdentifier(ident string) bool {
	for i, c := range ident {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return ident != "" && !IsKeyword(ident)
}

func NewEof(offset int) Token {
	return Token{
		Typ: Eof,
		Val: "",
		Span: span.Span{
			Beg: uint(offset),
			End: uint(offset),
		},
	}
}
