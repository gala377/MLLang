package syntax

import (
	"bytes"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/gala377/MLLang/syntax/token"
)

type Position struct {
	Line   uint
	Column uint
}

type Lexer struct {
	reader   *bytes.Reader
	position Position
	offset   int
	ch       rune
	eof      bool

	curr token.Token
	peek token.Token
}

func NewLexer(source io.Reader) Lexer {
	text, err := io.ReadAll(source)
	if err != nil {
		panic("panic")
	}
	var lexer Lexer
	lexer.reader = bytes.NewReader(text)
	// Populate peek so that the first Next call
	// returns properly
	lexer.readRune()
	lexer.Next()
	return lexer
}

func (l *Lexer) Next() token.Token {
	l.curr = l.peek
	l.moveToNextTok()
	return l.curr
}

func (l *Lexer) Current() token.Token {
	return l.curr
}

func (l *Lexer) Peek() token.Token {
	return l.peek
}

func (l *Lexer) moveToNextTok() {
	if l.position.Column > 1 {
		l.skipSpaces()
	}
	if l.eof {
		l.peek = token.NewEof(l.offset)
		return
	}
	l.peek = l.scanNextToken()
}

func (l *Lexer) scanNextToken() token.Token {
	var tok token.Token
	tok.Span.Beg = uint(l.offset - 1)
	ch := l.ch
	switch {
	case ch == '\n':
		tok.Typ = token.NewLine
		tok.Val = string(ch)
		l.readRune()
	case unicode.IsSpace(ch):
		tok.Typ = token.Indent
		tok.Val = l.scanIndent()
	case unicode.IsLetter(ch) || ch == '_':
		val := l.scanIdentifier()
		tok.Typ = token.Lookup(val)
		tok.Val = val
	case unicode.IsDigit(ch):
		// scan digit
	default:
		log.Fatalf("[%v:%v] Unknown character %v", l.position.Line, l.position.Column, ch)
	}
	tok.Span.End = uint(l.offset - 1)
	return tok
}

func (l *Lexer) readRune() rune {
	r, size, err := l.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			l.eof = true
			l.ch = -1
			// set offset past eof
			l.offset = int(l.reader.Size() + 1)
			return -1
		}
		log.Fatalf("[%v:%v] Could not decode character %v", l.position.Line, l.position.Column, err)
	}
	l.movePositionByRune(r)
	l.offset += size
	l.ch = r
	return r
}

func (l *Lexer) movePositionByRune(current rune) {
	if current == '\n' {
		l.position = Position{Line: l.position.Line + 1, Column: 0}
	} else {
		l.position.Column += 1
	}
}

func (l *Lexer) skipSpaces() {
	ch := l.ch
	for !l.eof && unicode.IsSpace(ch) && ch != '\n' {
		ch = l.readRune()
	}
}

func (l *Lexer) scanIndent() string {
	count := 0
	ch := l.ch
	for !l.eof && unicode.IsSpace(ch) && ch != '\n' {
		count += 1
		ch = l.readRune()
	}
	return strconv.Itoa(count)
}
func (l *Lexer) scanNumber() {}
func (l *Lexer) scanIdentifier() string {
	var b strings.Builder
	ch := l.ch
	for !l.eof && (unicode.IsLetter(ch) || ch == '_' || unicode.IsNumber(ch)) {
		b.WriteRune(ch)
		ch = l.readRune()
	}
	return b.String()
}
func (l *Lexer) scanStringLit() {}
