package syntax

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/gala377/MLLang/syntax/ast"
	"github.com/gala377/MLLang/syntax/span"
	"github.com/gala377/MLLang/syntax/token"
)

type parseExprFn = func() (ast.Expr, bool)
type SyntaxError struct {
	pos span.Span
	msg string
}

type Parser struct {
	l          *Lexer
	errors     []SyntaxError
	curr       token.Token
	indents    []int
	statements [token.Eof]parseExprFn
}

func NewParser(source io.Reader) *Parser {
	log.Println("Creating new parser")
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
	p.statements = [token.Eof]parseExprFn{
		token.Do:    p.parseLambda,
		token.While: p.parseWhile,
		token.If:    p.parseIf,
		token.Else: func() (ast.Expr, bool) {
			p.error(p.position(), p.position(), "else expected only after while")
			p.recover()
			return nil, false
		},
	}
	return &p
}

func (p *Parser) Parse() []ast.Node {
	log.Println("Parse")
	nodes := make([]ast.Node, 0)
	for !p.eof() {
		log.Println("Parse top level loop")
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
	log.Println("Eof")
	return nodes
}

func (p *Parser) parseTopLevelDecl() (ast.Decl, bool) {
	log.Println("Parsing top level decl")
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
	log.Println("Parse fn decl")
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
	body, ok := p.parseBlock()
	if !ok {
		return nil, false
	}
	if body == nil {
		if t := p.match(token.Assignment); t == nil {
			p.error(beg, p.position(), "expected colon or assignment in function definition")
			p.recover()
			return nil, false
		} else {
			ebody, ok := p.parseExpr()
			if !ok {
				return nil, false
			}
			if ebody == nil {
				p.error(beg, p.position(), "expected expression as a function body")
				p.recover()
				return nil, false
			}
			fbody = ebody
		}
	} else {
		fbody = body
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
	log.Println("Parsing val decl")
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
	log.Println("Parse top level expr")
	return p.parseExpr()
}

func (p *Parser) parseExpr() (ast.Expr, bool) {
	t := p.curr
	parseStmt := p.statements[t.Typ]
	if parseStmt == nil {
		return p.parseFunctionApp()
	}
	res, ok := parseStmt()
	if ok && res == nil {
		log.Printf("Could not parse the stmt, tok is %s", token.IdToString(p.curr.Typ))
		panic("chosen parse stmt function could not parse a stmt")
	}
	return res, ok
}

func (p *Parser) parseStmtExpression() (ast.Expr, bool) {
	t := p.curr
	parseStmt := p.statements[t.Typ]
	if parseStmt == nil {
		res, ok := p.parseFunctionApp()
		p.match(token.NewLine)
		return res, ok
	}
	res, ok := parseStmt()
	if ok && res == nil {
		log.Printf("Could not parse the stmt, tok is %s", token.IdToString(p.curr.Typ))
		panic("chosen parse stmt function could not parse a stmt")
	}
	return res, ok
}

func (p *Parser) parseBlock() (*ast.Block, bool) {
	log.Println("Parsing block")
	beg := p.position()
	if t := p.match(token.Colon); t == nil {
		return nil, true
	}
	if t := p.match(token.NewLine); t == nil {
		p.error(beg, p.position(), "expected new line for a block")
		p.recoverWithTokens(token.NewLine)
		return nil, false
	}
	indent, err := p.pushNextIndent()
	if err != nil {
		p.error(beg, p.position(), "expected block instructions to be indented")
		p.recover()
		return nil, false
	}
	defer p.popIndent(indent)
	parseExpr := func() (ast.Expr, bool) {
		log.Printf("%d running wrapped parse", indent)
		if t := p.matchIndent(indent); !t {
			return nil, true
		}
		return p.parseStmtExpression()
	}
	exprs := []ast.Node{}
	var e ast.Expr = nil
	ok := true
	for e, ok = parseExpr(); e != nil && ok; e, ok = parseExpr() {
		exprs = append(exprs, e)
	}
	if !ok {
		return nil, false
	}
	span := span.NewSpan(beg, p.position())
	node := ast.Block{
		Span:  &span,
		Instr: exprs,
	}
	return &node, true
}

func (p *Parser) parseFunctionApp() (ast.Expr, bool) {
	log.Println("Parse fn app")
	beg := p.position()
	fn, ok := p.parsePrimaryExpr()
	if !ok {
		return nil, false
	}
	arg, ok := p.parsePrimaryExpr()
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
		arg, ok = p.parsePrimaryExpr()
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

func (p *Parser) parsePrimaryExpr() (ast.Expr, bool) {
	log.Println("Parse primary expression")
	beg := p.position()
	tok := p.curr
	log.Printf("Token is %s\n", token.IdToString(tok.Typ))
	switch tok.Typ {
	case token.LParen:
		p.bump()
		node, ok := p.parseExpr()
		if !ok {
			return nil, false
		}
		t := p.match(token.RParen)
		if node == nil && t != nil {
			log.Println("Empty tuple")
			return p.emptyTuple(beg), true
		}
		if t == nil {
			log.Println("Parsing tuple")
			if t = p.match(token.Comma); t == nil {
				p.error(beg, p.position(), "missing closing parenthesis ')'")
			} else {
				// parsing tuple constant
				return p.parseTupleTail(beg, node)
			}
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
		log.Println("Not a primary")
		return nil, true
	}
}

func (p *Parser) parseWhile() (ast.Expr, bool) {
	beg := p.position()
	log.Printf("Parsing while, curr token is %s", token.IdToString(p.curr.Typ))
	if t := p.match(token.While); t == nil {
		return nil, true
	}
	cond, ok := p.parseExpr()
	if !ok {
		return nil, false
	}
	if cond == nil {
		p.error(beg, p.position(), "while expects an expression as its condition")
		p.recover()
		return nil, false
	}
	body, ok := p.parseBlock()
	if !ok {
		return nil, false
	}
	if body == nil {
		p.error(beg, p.position(), "while expects a block as its body")
		p.recover()
		return nil, false
	}
	span := span.NewSpan(beg, p.position())
	node := ast.WhileExpr{
		Span: &span,
		Cond: cond,
		Body: body,
	}
	return &node, true
}

func (p *Parser) parseIf() (ast.Expr, bool) {
	beg := p.position()
	if t := p.match(token.If); t == nil {
		return nil, true
	}
	cond, ok := p.parseExpr()
	if !ok {
		return nil, false
	}
	if cond == nil {
		p.error(beg, p.position(), "if expects an expression as its condition")
		p.recover()
		return nil, false
	}
	body, ok := p.parseBlock()
	if !ok {
		return nil, false
	}
	if body == nil {
		p.error(beg, p.position(), "if expects a block as its body")
		p.recover()
		return nil, false
	}
	elseb, ok := p.parseElse()
	if !ok {
		return nil, false
	}
	span := span.NewSpan(beg, p.position())
	node := ast.IfExpr{
		Span:       &span,
		Cond:       cond,
		IfBranch:   body,
		ElseBranch: elseb,
	}
	return &node, true
}

func (p *Parser) parseElse() (ast.Expr, bool) {
	if t := p.match(token.Else); t == nil {
		return nil, true
	}
	if p.curr.Typ == token.If {
		return p.parseIf()
	}
	return p.parseBlock()
}

func (p *Parser) parseLambda() (ast.Expr, bool) {
	log.Println("Parsing lambda")
	beg := p.position()
	if t := p.match(token.Do); t == nil {
		return nil, true
	}
	args := []ast.FuncDeclArg{}
	if t := p.match(token.Pipe); t != nil {
		log.Println("Parsing lambda arguments")
		for pt := p.match(token.Pipe); pt == nil; pt = p.match(token.Pipe) {
			arg := p.parseIdentifier()
			if arg == nil {
				p.error(beg, p.position(), "Lambda argument has to be an identifier")
				p.recoverWithTokens(token.Pipe)
				continue
			}
			a := ast.FuncDeclArg{
				Span: arg.Span,
				Name: arg.Name,
			}
			log.Printf("Parsed parameter %s", a.Name)
			args = append(args, a)
		}
	}
	log.Println("Parsed lambda arguments")
	var body ast.Expr
	if t := p.match(token.Colon); t != nil {
		b, ok := p.parseBlock()
		if !ok {
			return nil, false
		}
		if b == nil {
			p.error(beg, p.position(), "Expected block as a lambda body")
			p.recover()
			return nil, false
		}
		body = b
	} else if t := p.match(token.Operator); t != nil && token.IsArrow(t) {
		e, ok := p.parseExpr()
		if !ok {
			return nil, false
		}
		if e == nil {
			p.error(beg, p.position(), "Expected expr as a lambda body")
			p.recover()
			return nil, false
		}
		body = e
	} else {
		log.Printf("Expected -> or block got: %s\n", p.curr.Val)
		p.error(beg, p.position(), "Expected -> or a block following lambda argument list")
		p.recover()
		return nil, false
	}
	span := span.NewSpan(beg, p.position())
	node := ast.LambdaExpr{
		Span: &span,
		Args: args,
		Body: body,
	}
	return &node, true

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

func (p *Parser) parseTupleTail(beg span.Position, first ast.Expr) (*ast.TupleConst, bool) {
	log.Println("Parsing tuple tail")
	vals := []ast.Expr{first}
	for {
		log.Println("Parsing tuple val")
		e, ok := p.parseExpr()
		if e == nil {
			if !ok {
				p.error(beg, p.position(), "Expected expression in tuple constant")
				p.recoverWithTokens(token.Comma, token.NewLine, token.RParen)
				continue
			}
			break
		}
		log.Printf("Parsed tuple val %s\n", e.String())
		vals = append(vals, e)
		if t := p.match(token.Comma); t == nil {
			break
		}
	}
	log.Println("Endend tuple parsing")
	if t := p.match(token.RParen); t == nil {
		p.error(beg, p.position(), "Expected ) to close a tuple literal")
	}
	span := span.NewSpan(beg, p.position())
	tuple := &ast.TupleConst{
		Span: &span,
		Vals: vals,
	}
	return tuple, true
}

func (p *Parser) emptyTuple(beg span.Position) *ast.TupleConst {
	span := span.NewSpan(beg, p.position())
	return &ast.TupleConst{
		Span: &span,
		Vals: []ast.Expr{},
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

func (p *Parser) matchIndent(n int) bool {
	t := p.l.Current()
	if t.Typ != token.Indent {
		return false
	}
	val, err := strconv.Atoi(t.Val)
	if err != nil {
		log.Panicf("could not parse indentations value, %s", t.Val)
	}
	if val != n {
		log.Printf("Indent not matching %d != %d", n, val)
		return false
	}
	p.bump()
	return true
}

func (p *Parser) bump() {
	p.curr = p.l.Next()
}

func (p *Parser) pushNextIndent() (int, error) {
	t := p.l.Current()
	if t.Typ != token.Indent {
		return 0, errors.New("no indentation to parse")
	}
	val, err := strconv.Atoi(t.Val)
	if err != nil {
		log.Panicf("indentation value could not be retrieved %s", t.Val)
	}
	curr := p.indents[len(p.indents)-1]
	if val <= curr {
		return 0, fmt.Errorf("indentation expected to be higher: %d <= %d", val, curr)
	}
	p.indents = append(p.indents, val)
	return val, nil
}

func (p *Parser) popIndent(n int) {
	curr := p.indents[len(p.indents)-1]
	if curr != n {
		log.Panicf("trying to pop wrong indentation %d != %d", n, curr)
	}
	p.indents = p.indents[:len(p.indents)-1]
}

func (p *Parser) error(beg, end span.Position, msg string) {
	p.errors = append(p.errors, SyntaxError{
		pos: span.NewSpan(beg, end),
		msg: msg,
	})
}

func (p *Parser) recover() {
	p.recoverWithTokens()
}

func (p *Parser) recoverWithTokens(rtt ...token.Id) {
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
	return p.l.curr.Typ == token.Eof
}

func (p *Parser) position() span.Position {
	return p.l.position
}
