package syntax

import (
	"io"
	"log"
	"strconv"

	"github.com/gala377/MLLang/syntax/ast"
	"github.com/gala377/MLLang/syntax/span"
	"github.com/gala377/MLLang/syntax/token"
)

type SyntaxError struct {
	pos span.Span
	msg string
}

type Parser struct {
	l       *Lexer
	errors  []SyntaxError
	curr    token.Token
	indents []int
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
	p.curr = l.Next()
	p.indents = []int{0}
	return p
}

func (p *Parser) Parse() []ast.Node {
	nodes := make([]ast.Node, 0)
	for !p.eof() {
		var n ast.Node
		var ok bool
		n, ok = p.parseTopLevelDecl()
		if n != nil || !ok {
			if n != nil {
				nodes = append(nodes, n)
			}
			continue
		}
		n, ok = p.parseTopLevelExpr()
		if n != nil || !ok {
			if n != nil {
				nodes = append(nodes, n)
			}
			continue
		}
		bp := p.position()
		p.recover()
		p.error(bp, p.position(), "expected declaration or expression")
	}
	return nodes
}

func (p *Parser) parseTopLevelDecl() (ast.Decl, bool) {
	fnode, ok := p.parseFnDecl()
	if fnode != nil || !ok {
		return fnode, ok
	}
	vnode, ok := p.parseValDecl()
	if vnode != nil || !ok {
		return vnode, ok
	}
	return nil, true
}

func (p *Parser) parseFnDecl() (*ast.FuncDecl, bool) {
	return nil, true
}

func (p *Parser) parseValDecl() (*ast.GlobalValDecl, bool) {
	beg := p.position()
	if t := p.match(token.Val); t == nil {
		return nil, true
	}
	name := p.match(token.Identifier)
	if name == nil {
		p.error(beg, p.position(), "expected identifier in variable declaration")
		p.recover()
		return nil, false
	}
	if t := p.match(token.Assignment); t == nil {
		p.error(beg, p.position(), "expected '=' operator in variable declaration")
		p.recover()
		return nil, false
	}
	expr, ok := p.parseExpr()
	span := span.NewSpan(beg, p.position())
	node := ast.GlobalValDecl{
		Span: &span,
		Name: name.Val,
		Rhs:  expr,
	}
	return &node, ok
}

func (p *Parser) parseTopLevelExpr() (ast.Expr, bool) {
	return p.parseExpr()
}

func (p *Parser) parseExpr() (ast.Expr, bool) {
	return nil, true
}

func (p *Parser) parseFunctionApp() (*ast.FuncApplication, bool) {
	return nil, true
}

func (p *Parser) parsePrimaryExpression() (ast.Expr, bool) {
	beg := p.position()
	tok := p.curr
	switch tok.Typ {
	case token.LParen:
		p.bump()
		node, ok := p.parseExpr()
		t := p.match(token.RParen)
		if !ok {
			// todo: how could we recover from that?
			return nil, false
		}
		if t == nil {
			p.error(beg, p.position(), "missing closing parenthesis ')'")
		}
		return node, true
	case token.Integer:
		p.bump()
		var node ast.IntConst
		node.Span = p.curr.Span
		val, err := strconv.Atoi(tok.Val)
		if err != nil {
			log.Panicf("unreachable: could not convert int token value to integer: %s", tok.Val)
		}
		node.Val = val
		return &node, true
	case token.Float:
		p.bump()
		var node ast.FloatConst
		node.Span = p.curr.Span
		val, err := strconv.ParseFloat(tok.Val, 64)
		if err != nil {
			log.Panicf("unreachable: could not convert float token value to float64: %s", tok.Val)
		}
		node.Val = val
		return &node, true
	case token.String:
		p.bump()
		var node ast.StringConst
		node.Span = p.curr.Span
		node.Val = tok.Val
		return &node, true
	case token.LSquareParen, token.LBracket:
		panic("not implemented")
	default:
		return nil, true
	}
}

func (p *Parser) match(typ token.Id) *token.Token {
	if p.curr.Typ == typ {
		tok := p.curr
		p.bump()
		return &tok
	}
	return nil
}

func (p *Parser) bump() {
	p.curr = p.l.Next()
}

func (p *Parser) error(beg, end span.Position, msg string) {
	p.errors = append(p.errors, SyntaxError{
		pos: span.NewSpan(beg, end),
		msg: msg,
	})
}

func (p *Parser) recover() {
	p.recoverWithTokens([]token.Id{})
}

func (p *Parser) recoverWithTokens(rtt []token.Id) {
	defer p.l.UnsetMode(skipErrorReporting)
	p.l.SetMode(skipErrorReporting)
	for t := p.l.curr; !p.eof() && !isRecoveryToken(t, rtt); t = p.l.Next() {
		if t.Typ == token.NewLine && p.l.Peek().Typ != token.Indent {
			break
		}
	}
	p.curr = p.l.curr
}

func isRecoveryToken(t token.Token, rtt []token.Id) bool {
	for _, r := range rtt {
		if t.Typ == r {
			return true
		}
	}
	return false
}

func (p *Parser) eof() bool {
	return p.l.eof
}

func (p *Parser) position() span.Position {
	return p.l.position
}
