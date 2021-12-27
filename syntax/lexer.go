package syntax

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/gala377/MLLang/syntax/span"
	"github.com/gala377/MLLang/syntax/token"
)

const controlChars string = "'\"(){}[]+-=/<>!~@#$%^&*|,;:`"
const operatorChars string = "+=-\\/<>!@#$%^&*~?"

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
	Mode  uint64
	Lexer struct {
		reader   *bytes.Reader
		err      ErrorHandler
		position span.Position
		offset   int
		ch       rune
		eof      bool
		mode     Mode

		curr token.Token
		peek token.Token
	}

	ErrorHandler = func(beg, end span.Position, msg string)
)

const (
	returnComments Mode = 1 << iota
	skipErrorReporting
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
		l.peek = token.NewEof(l.position)
		return
	}
	l.peek = l.scanNextToken()
}

func (l *Lexer) scanNextToken() token.Token {
	tok := l.newToken()
	ch := l.ch
	bpos := l.position
	var err error
	switch {
	case ch == '\n':
		tok.Typ = token.NewLine
		tok.Val = string(ch)
		l.readRune()
	case unicode.IsSpace(ch):
		tok.Typ = token.Indent
		tok.Val = l.scanIndent()
	case isValidFirstIdentifierChar(ch):
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
	case ch == ';':
		tok.Val = l.scanComment()
		tok.Typ = token.Comment
		if !l.GetMode(returnComments) {
			if l.eof {
				return token.NewEof(l.position)
			}
			return l.scanNextToken()
		}
	case ch == ':':
		nch := l.readRune()
		if isValidFirstIdentifierChar(nch) {
			val := l.scanIdentifier()
			if token.Lookup(val) != token.Identifier {
				err = fmt.Errorf("keyword as infix indentifier is illegal")
			}
			tok.Typ = token.InfixIdentifier
			tok.Val = val
		} else {
			tok.Typ = token.Colon
			tok.Val = token.IdToString(tok.Typ)
		}
	default:
		known := true
		switch ch {
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
		case '|':
			// todo: if another opearator character follows then it's
			// a operator. So for exampl |> |+|
			tok.Typ = token.Pipe
		case '.':
			tok.Typ = token.Access
		case '`':
			tok.Typ = token.Quote
		default:
			err = fmt.Errorf("unknown character %s", string(ch))
			known = false
		}
		if known {
			tok.Val = token.IdToString(tok.Typ)
			l.readRune()
		}
	}
	if err != nil {
		tok.Typ = token.Error
		if !l.GetMode(skipErrorReporting) {
			l.err(bpos, l.position, err.Error())
		}
		recovered := l.recover()
		tok.Val += recovered
	}
	tok.Span.End = l.position
	tok.Span.End.Offset = uint(l.offset - 1)
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
	l.offset += size
	l.movePositionByRune(r)
	l.ch = r
	return r
}

func (l *Lexer) movePositionByRune(current rune) {
	if current == '\n' {
		l.position = span.Position{Line: l.position.Line + 1, Column: 0, Offset: 0}
	} else {
		l.position.Column += 1
	}
	l.position.Offset = uint(l.offset)
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
		if ch == '\\' {
			switch l.readRune() {
			case 'n':
				b.WriteRune('\n')
			case '\\':
				b.WriteRune('\\')
			case 't':
				b.WriteRune('\t')
			default:
				l.recover()
				return "", fmt.Errorf("unknown escape character in string literal %v", l.ch)
			}
		} else {
			b.WriteRune(ch)
		}
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

func (l *Lexer) scanComment() string {
	var b strings.Builder
	ch := l.readRune()
	for !l.eof && ch != '\n' {
		b.WriteRune(ch)
		ch = l.readRune()
	}
	return b.String()
}

func (l *Lexer) recover() string {
	var b strings.Builder
	ch := l.ch
	cont := isControl(ch)
	for !(unicode.IsSpace(ch) || cont || l.eof) {
		b.WriteRune(ch)
		ch = l.readRune()
		cont = isControl(ch)
	}
	return b.String()
}

func (l *Lexer) newToken() token.Token {
	var t token.Token
	span := span.Span{
		Beg: l.position,
		End: l.position,
	}
	span.Beg.Offset = uint(l.offset - 1)
	t.Span = &span
	return t
}

func (l *Lexer) SetMode(flag Mode) {
	l.mode |= flag
}

func (l *Lexer) UnsetMode(flag Mode) {
	l.mode &= ^flag
}

func (l *Lexer) GetMode(flag Mode) bool {
	return (l.mode & flag) > 0
}

func isControl(ch rune) bool {
	_, ok := controlCharsSet[ch]
	return ok
}

func isValidFirstIdentifierChar(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isValidInIdentifier(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '?' || unicode.IsNumber(ch)
}

func isValidInOperator(ch rune) bool {
	_, ok := operatorCharsSet[ch]
	return ok
}
