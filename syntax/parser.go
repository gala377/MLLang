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
	p.errors = make([]SyntaxError, 0)
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
		switch p.curr.Typ {
		case token.Eof:
			continue
		case token.NewLine:
			p.bump()
		case token.Indent:
			p.error(p.curr.Span.Beg, p.position(), "Unexpected indentantion at the top level")
			p.recover()
		default:
			bp := p.position()
			p.recover()
			p.error(bp, p.position(), "expected declaration or expression")
		}
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
	beg := p.position()
	if t := p.match(token.Fn); t == nil {
		return nil, true
	}
	name := p.parseIdentifier()
	if name == nil {
		p.error(beg, p.position(), "expected function name")
		p.recover()
		return nil, false
	}
	args := []*ast.Identifier{}
	for arg := p.parseIdentifier(); arg != nil; arg = p.parseIdentifier() {
		args = append(args, arg)
	}
	var fbody ast.Expr
	if t := p.match(token.Colon); t != nil {
		block, ok := p.parseBlock()
		if !ok {
			return nil, false
		}
		if block == nil {
			p.error(beg, p.position(), "expected a block as a function body")
			p.recover()
			return nil, false
		}
		fbody = block
	} else if t := p.match(token.Assignment); t != nil {
		body, ok := p.parseExpr()
		if !ok {
			return nil, false
		}
		if body == nil {
			p.error(beg, p.position(), "expected expression as a function body")
			p.recover()
			return nil, false
		}
		fbody = body
	} else {
		p.error(beg, p.position(), "expected colon or assignment in function definition")
		p.recover()
		return nil, false
	}
	span := span.NewSpan(beg, p.position())
	fargs := []ast.FuncDeclArg{}
	for _, arg := range args {
		fargs = append(fargs, ast.FuncDeclArg{
			Span: arg.Span,
			Name: arg.Name,
		})
	}
	fn := ast.FuncDecl{
		Span: &span,
		Name: name.Name,
		Args: fargs,
		Body: fbody,
	}
	return &fn, true
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

func (p *Parser) parseBlock() (ast.Expr, bool) {
	panic("Parse block not implemented")
}

func (p *Parser) parseExpr() (ast.Expr, bool) {
	return p.parseFunctionApp()
}

func (p *Parser) parseFunctionApp() (ast.Expr, bool) {
	beg := p.position()
	fn, ok := p.parsePrimaryExpression()
	if !ok {
		return nil, false
	}
	arg, ok := p.parsePrimaryExpression()
	if !ok {
		return nil, false
	}
	if arg == nil {
		// not a function call, just normal expression
		return fn, true
	}
	args := []ast.Expr{}
	for arg != nil {
		args = append(args, arg)
		arg, ok = p.parsePrimaryExpression()
		if !ok {
			return nil, false
		}
	}
	span := span.NewSpan(beg, p.position())
	node := ast.FuncApplication{
		Span:   &span,
		Callee: fn,
		Args:   args,
	}
	return &node, true
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
		node.Span = tok.Span
		val, err := strconv.Atoi(tok.Val)
		if err != nil {
			log.Panicf("unreachable: could not convert int token value to integer: %s", tok.Val)
		}
		node.Val = val
		return &node, true
	case token.Float:
		p.bump()
		var node ast.FloatConst
		node.Span = tok.Span
		val, err := strconv.ParseFloat(tok.Val, 64)
		if err != nil {
			log.Panicf("unreachable: could not convert float token value to float64: %s", tok.Val)
		}
		node.Val = val
		return &node, true
	case token.String:
		p.bump()
		var node ast.StringConst
		node.Span = tok.Span
		node.Val = tok.Val
		return &node, true
	case token.Identifier:
		p.bump()
		var node ast.Identifier
		node.Span = tok.Span
		node.Name = tok.Val
		return &node, true
	case token.LSquareParen, token.LBracket:
		panic("not implemented")
	default:
		return nil, true
	}
}

func (p *Parser) parseIdentifier() *ast.Identifier {
	t := p.match(token.Identifier)
	if t == nil {
		return nil
	}
	return &ast.Identifier{
		Span: t.Span,
		Name: t.Val,
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
	p.bump()
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
	return p.curr.Typ == token.Eof
}

func (p *Parser) position() span.Position {
	return p.l.position
}
