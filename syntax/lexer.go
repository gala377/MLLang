package syntax

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/gala377/MLLang/syntax/token"
)

const controlChars string = "'\"(){}[]+-=/<>!~@#$%^&*|,;:`"

var controlCharsSet map[rune]bool

func Init() {
	for _, char := range controlChars {
		controlCharsSet[char] = true
	}
}

type (
	Position struct {
		Line   uint
		Column uint
	}

	Lexer struct {
		reader   *bytes.Reader
		err      ErrorHandler
		position Position
		offset   int
		ch       rune
		eof      bool

		curr token.Token
		peek token.Token
	}

	ErrorHandler = func(pos Position, msg string)
)

func NewLexer(source io.Reader, report ErrorHandler) Lexer {
	text, err := io.ReadAll(source)
	if err != nil {
		panic("panic")
	}
	var lexer Lexer
	lexer.err = report
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
		val, float := l.scanNumber()
		tok.Val = val
		tok.Typ = token.Integer
		if float {
			tok.Typ = token.Float
		}
	case ch == '"' || ch == '\'':
		tok.Typ = token.String
		tok.Val = l.scanStringLit()
	default:
		msg := fmt.Sprintf("unknown character %v", ch)
		l.err(l.position, msg)
		l.recover()
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
		// We can assume that if we cannot decode the character
		// then the scanned file can be assumed to be garbage so
		// no need for reporting. Just abort.
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
func (l *Lexer) scanNumber() (val string, isfloat bool) {
	var b strings.Builder
	ch := l.ch
	if ch == '0' {
		b.WriteRune(ch)
		switch ch = l.readRune(); {
		case unicode.IsSpace(ch) || isControl(ch):
			val = "0"
		case ch == '.':
			isfloat = true
			b.WriteRune(ch)
			l.readRune()
			l.scanNumbersFractionPart(&b)
			val = b.String()
		}
		return
	}
	for unicode.IsDigit(ch) {
		b.WriteRune(ch)
		ch = l.readRune()
	}
	if ch == '.' {
		isfloat = true
		b.WriteRune(ch)
		l.readRune()
		l.scanNumbersFractionPart(&b)
	} else if unicode.IsLetter(ch) {
		b.WriteRune(ch)
		msg := fmt.Sprintf("Expected a number but got: '%v' which is not a number", b.String())
		l.err(l.position, msg)
		l.recover()
	}
	val = b.String()
	return
}

func (l *Lexer) scanNumbersFractionPart(b *strings.Builder) {
	ch := l.ch
	for unicode.IsNumber(ch) {
		b.WriteRune(ch)
		ch = l.readRune()
	}
	if unicode.IsLetter(ch) || ch == '.' {
		b.WriteRune(ch)
		msg := fmt.Sprintf("expected a number but got '%v'", b.String())
		l.err(l.position, msg)
		l.recover()
	}
}

func (l *Lexer) scanIdentifier() string {
	var b strings.Builder
	ch := l.ch
	for !l.eof && isValidInIdentifier(ch) {
		b.WriteRune(ch)
		ch = l.readRune()
	}
	return b.String()
}
func (l *Lexer) scanStringLit() string {
	var b strings.Builder
	quote := l.ch
	ch := l.readRune()
	for !l.eof && ch != quote && ch != '\n' {
		log.Println("scanning", string(quote), string(ch), ch == quote)
		b.WriteRune(ch)
		ch = l.readRune()
	}
	if l.eof {
		l.err(l.position, "expected string closing quote but got eof")
	} else if ch != quote {
		l.err(l.position, "unclosed string")
	} else {
		l.readRune()
	}
	return b.String()
}

func (l *Lexer) recover() {
	ch := l.ch
	cont := isControl(ch)
	for !(unicode.IsSpace(ch) || cont) {
		ch = l.readRune()
		cont = isControl(ch)
	}
}

func isControl(ch rune) bool {
	_, ok := controlCharsSet[ch]
	return ok
}

func isValidInIdentifier(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '?' || unicode.IsNumber(ch)
}
