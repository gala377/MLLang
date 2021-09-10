package syntax

import (
	"bytes"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"
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

	curr Token
	peek Token
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

func (l *Lexer) Next() Token {
	l.curr = l.peek
	l.moveToNextTok()
	return l.curr
}

func (l *Lexer) Current() Token {
	return l.curr
}

func (l *Lexer) Peek() Token {
	return l.peek
}

func (l *Lexer) moveToNextTok() {
	if l.position.Column > 1 {
		l.skipSpaces()
	}
	if l.eof {
		l.peek = newEof(l.offset)
		return
	}
	l.peek = l.scanNextToken()
}

func (l *Lexer) scanNextToken() Token {
	var tok Token
	tok.span.Beg = uint(l.offset - 1)
	ch := l.ch
	switch {
	case ch == '\n':
		tok.typ = NewLine
		tok.val = string(ch)
		l.readRune()
	case unicode.IsSpace(ch):
		tok.typ = Indent
		tok.val = l.scanIndent()
	case unicode.IsLetter(ch) || ch == '_':
		val := l.scanIdentifier()
		tok.typ = Lookup(val)
		tok.val = val
	case unicode.IsDigit(ch):
		// scan digit
	default:
		log.Fatalf("[%v:%v] Unknown character %v", l.position.Line, l.position.Column, ch)
	}
	tok.span.End = uint(l.offset - 1)
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
