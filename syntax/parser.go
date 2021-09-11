package syntax

import (
	"go/ast"
	"io"

	"github.com/gala377/MLLang/syntax/span"
	"github.com/gala377/MLLang/syntax/token"
)

type SyntaxError struct {
	pos span.Span
	msg string
}

type Parser struct {
	l      *Lexer
	errors []SyntaxError
}

func NewParser(source io.Reader) Parser {
	var p Parser
	handler := func(beg, end span.Position, msg string) {
		p.errors = append(p.errors, SyntaxError{
			pos: span.NewSpan(beg, end),
			msg: msg,
		})
	}
	l := NewLexer(source, handler)
	p.l = &l
	return p
}

func (p *Parser) Parse() []ast.Node {
	nodes := make([]ast.Node, 0)
	for !p.eof() {
		n := p.parseTopLevelDecl()
		if n != nil {
			nodes = append(nodes, n)
			continue
		}
		n = p.parseTopLevelExpr()
		if n != nil {
			nodes = append(nodes, n)
			continue
		}
		bp := p.position()
		p.recover()
		p.error(bp, p.position(), "Expected declaration or expression")
	}
	return nodes
}

func (p *Parser) parseTopLevelDecl() ast.Node {
	return nil
}

func (p *Parser) parseTopLevelExpr() ast.Node {
	return nil
}

func (p *Parser) error(beg, end span.Position, msg string) {
	p.errors = append(p.errors, SyntaxError{
		pos: span.NewSpan(beg, end),
		msg: msg,
	})
}

func (p *Parser) recover() {
	defer p.l.UnsetMode(skipErrorReporting)
	p.l.SetMode(skipErrorReporting)
	for t := p.l.Next(); !p.eof() && !isRecoveryToken(t); t = p.l.Next() {
	}
}

func (p *Parser) eof() bool {
	return p.l.eof
}

func (p *Parser) position() span.Position {
	return p.l.position
}

func isRecoveryToken(t token.Token) bool {
	switch t.Typ {
	case token.NewLine, token.Eof, token.RBracket, token.RParen, token.RSquareParen, token.Comma, token.Colon:
		return true
	case token.Fn, token.Val, token.If, token.While, token.Let, token.Else:
		return true
	default:
		return false
	}
}
