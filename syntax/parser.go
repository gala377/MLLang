package syntax

import (
	"go/ast"
	"io"

	"github.com/gala377/MLLang/syntax/span"
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

func (p *Parser) eof() bool {
	return p.l.eof
}
