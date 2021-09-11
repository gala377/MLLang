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
const operatorChars string = "+=-\\/<>;!@#$%^&*|~?"

var controlCharsSet map[rune]bool
var operatorCharsSet map[rune]bool

func init() {
	controlCharsSet = make(map[rune]bool)
	for _, char := range controlChars {
		controlCharsSet[char] = true
	}
	operatorCharsSet = make(map[rune]bool)
	for _, char := range operatorChars {
		operatorCharsSet[char] = true
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
	var err error
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
		val, float, serr := l.scanNumber()
		tok.Val = val
		tok.Typ = token.Integer
		err = serr
		if float {
			tok.Typ = token.Float
		}
	case ch == '"' || ch == '\'':
		tok.Typ = token.String
		tok.Val, err = l.scanStringLit()
	case isValidInOperator(ch):
		val := l.scanOperator()
		tok.Typ = token.LookupOperator(val)
		tok.Val = val
	default:
		read_needed := true
		switch ch {
		case ':':
			tok.Typ = token.Colon
		case '(':
			tok.Typ = token.LParen
		case ')':
			tok.Typ = token.RParen
		case '[':
			tok.Typ = token.LSquareParen
		case ']':
			tok.Typ = token.RSquareParen
		case '{':
			tok.Typ = token.LBracket
		case '}':
			tok.Typ = token.RBracket
		case ',':
			tok.Typ = token.Comma
		default:
			err = fmt.Errorf("unknown character %s", string(ch))
			read_needed = false
		}
		if read_needed {
			l.readRune()
		}
		tok.Val = token.IdToString(tok.Typ)
	}
	tok.Span.End = uint(l.offset - 1)
	if err != nil {
		tok.Typ = token.Error
		l.err(l.position, err.Error())
		l.recover()
	}
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
func (l *Lexer) scanNumber() (val string, isfloat bool, err error) {
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
			err = l.scanNumbersFractionPart(&b)
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
		err = l.scanNumbersFractionPart(&b)
	} else if unicode.IsLetter(ch) {
		b.WriteRune(ch)
		err = fmt.Errorf("expected a number but got: '%v' which is not a number", b.String())
	}
	val = b.String()
	return
}

func (l *Lexer) scanNumbersFractionPart(b *strings.Builder) error {
	ch := l.ch
	for unicode.IsNumber(ch) {
		b.WriteRune(ch)
		ch = l.readRune()
	}
	if unicode.IsLetter(ch) || ch == '.' {
		b.WriteRune(ch)
		return fmt.Errorf("expected a number but got '%v'", b.String())
	}
	return nil
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
func (l *Lexer) scanStringLit() (string, error) {
	var b strings.Builder
	quote := l.ch
	ch := l.readRune()
	for !l.eof && ch != quote && ch != '\n' {
		b.WriteRune(ch)
		ch = l.readRune()
	}
	var err error
	if l.eof {
		err = fmt.Errorf("expected string closing quote but got eof")
	} else if ch != quote {
		err = fmt.Errorf("unclosed string")
	} else {
		l.readRune()
	}
	return b.String(), err
}

func (l *Lexer) scanOperator() string {
	var b strings.Builder
	ch := l.ch
	for isValidInOperator(ch) {
		b.WriteRune(ch)
		ch = l.readRune()
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

func isValidInOperator(ch rune) bool {
	_, ok := operatorCharsSet[ch]
	return ok
}
