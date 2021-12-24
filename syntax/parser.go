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
type parseStmtFn = func() (ast.Stmt, bool)
type SyntaxError struct {
	pos span.Span
	msg string
}

func (s SyntaxError) SourceLoc() span.Span {
	return s.pos
}

func (s SyntaxError) Error() string {
	return s.msg
}

type Parser struct {
	l                   *Lexer
	errors              []SyntaxError
	curr                token.Token
	indents             []int
	stmtSpecialForms    [token.Eof + 1]parseStmtFn
	exprSpecialForms    [token.Eof + 1]parseExprFn
	parseTrailingBlocks bool
	scope               *Scope
}

func NewParser(source io.Reader) *Parser {
	log.Println("==============Creating new parser===============")
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
	p.exprSpecialForms = [token.Eof + 1]parseExprFn{
		token.Do:     p.parseLambda,
		token.If:     p.parseIf,
		token.Handle: p.parseHandle,
		token.Resume: p.parseResume,
		token.Else: func() (ast.Expr, bool) {
			p.error(p.position(), p.position(), "else expected only after while")
			p.recover()
			return nil, false
		},
		token.With: func() (ast.Expr, bool) {
			p.error(p.position(), p.position(), "with expected only after handle")
			p.recover()
			return nil, false
		},
	}
	p.stmtSpecialForms = [token.Eof + 1]parseStmtFn{
		token.While:  p.parseWhile,
		token.Let:    p.parseValDecl,
		token.Return: p.parseReturn,
		token.Fn: func() (ast.Stmt, bool) {
			fn, ok := p.parseLocalFnDecl()
			if fn == nil || !ok {
				return nil, ok
			}
			decl := ast.ValDecl{
				Span: fn.Span,
				Name: fn.Name,
				Rhs:  fn,
			}
			p.scope.InsertVal(&decl)
			return &decl, true
		},
		token.Effect: func() (ast.Stmt, bool) {
			eff, ok := p.parseLocalEffectDecl()
			if eff == nil || !ok {
				return nil, ok
			}
			decl := ast.ValDecl{
				Span: eff.Span,
				Name: eff.Name,
				Rhs:  eff,
			}
			p.scope.InsertVal(&decl)
			return &decl, true
		},
	}
	p.parseTrailingBlocks = true
	p.scope = NewScope(nil)
	return &p
}

func (p *Parser) Errors() []SyntaxError {
	return p.errors
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
		n, ok = p.parseTopLevelStmt()
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
	fnode, ok := p.parseGlobalFnDecl()
	if fnode != nil || !ok {
		return fnode, ok
	}
	vnode, ok := p.parseGlobalValDecl()
	if vnode != nil || !ok {
		return vnode, ok
	}
	enode, ok := p.parseTopLevelEffectDecl()
	if enode != nil || !ok {
		return enode, ok
	}
	return nil, true
}

func (p *Parser) parseGlobalFnDecl() (*ast.FuncDecl, bool) {
	log.Println("Parse global fn decl")
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
	p.scope.Insert(name.Name)
	if !p.scope.IsGlobal() {
		panic("ICE: Expected global scope")
	}
	p.openScope()
	defer p.closeScope()

	args := []*ast.FuncDeclArg{}
	for arg := p.parseIdentifier(); arg != nil; arg = p.parseIdentifier() {
		farg := &ast.FuncDeclArg{Span: arg.Span, Name: arg.Name}
		args = append(args, farg)
		p.scope.InsertFuncArg(farg)
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
	fn := ast.FuncDecl{
		Span: &span,
		Name: name.Name,
		Args: args,
		Body: fbody,
	}
	return &fn, true
}

func (p *Parser) parseGlobalValDecl() (*ast.GlobalValDecl, bool) {
	log.Println("Parsing val decl")
	beg := p.position()
	if t := p.match(token.Let); t == nil {
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
	if !p.scope.IsGlobal() {
		panic("ICE: expected global scope")
	}
	p.scope.Insert(node.Name)
	return &node, ok
}

func (p *Parser) parseTopLevelEffectDecl() (*ast.EffectDecl, bool) {
	beg := p.position()
	if p.match(token.Effect) == nil {
		return nil, true
	}
	name := p.parseIdentifier()
	if name == nil {
		p.error(beg, p.position(), "Expected idnetifier as an effect name")
		return nil, false
	}
	span := span.NewSpan(beg, p.position())
	if !p.scope.IsGlobal() {
		panic("ICE: expected global scope")
	}
	p.scope.Insert(name.Name)
	return &ast.EffectDecl{
		Name: name.Name,
		Span: &span,
	}, true
}

func (p *Parser) parseTopLevelStmt() (ast.Stmt, bool) {
	log.Println("Parse top level expr")
	return p.parseStmt()
}

func (p *Parser) parseStmt() (ast.Stmt, bool) {
	t := p.curr
	parseSpecialForm := p.stmtSpecialForms[t.Typ]
	if parseSpecialForm == nil {
		lval, ok := p.parseExpr()
		if lval == nil {
			return nil, ok
		}
		var node ast.Stmt = &ast.StmtExpr{Expr: lval}
		if p.match(token.Assignment) != nil {
			if id, ok := lval.(*ast.Identifier); ok {
				p.tryLiftVar(id.Name)
			}
			rval, ok := p.parseExpr()
			if rval == nil || !ok {
				if ok {
					p.error(t.Span.Beg, p.position(), "expected expression after assigment operator")
					p.recover()
				}
				return nil, false
			}
			span := span.NewSpan(t.Span.Beg, p.position())
			node = &ast.Assignment{Span: &span, LValue: lval, RValue: rval}
		}
		p.match(token.NewLine)
		return node, ok
	}
	res, ok := parseSpecialForm()
	if ok && res == nil {
		log.Printf("Could not parse the stmt, tok is %s", token.IdToString(p.curr.Typ))
		panic("chosen parse stmt function could not parse a stmt")
	}
	return res, ok
}

func (p *Parser) parseExpr() (ast.Expr, bool) {
	t := p.curr
	parseSpecialForm := p.exprSpecialForms[t.Typ]
	if parseSpecialForm == nil {
		return p.parseBinaryExpression()
	}
	res, ok := parseSpecialForm()
	if ok && res == nil {
		log.Printf("Could not parse the expr, tok is %s", token.IdToString(p.curr.Typ))
		panic("chosen parse expr function could not parse an expr")
	}
	return res, ok
}

func (p *Parser) parseBinaryExpression() (ast.Expr, bool) {
	beg := p.position()
	fapp, ok := p.parseFunctionApp()
	if fapp == nil || !ok {
		return fapp, ok
	}
	applications := []ast.Expr{fapp}
	for p.match(token.Dollar) != nil {
		fapp, ok := p.parseFunctionApp()
		if fapp == nil || !ok {
			p.error(beg, p.position(), "expected expression after binary operator")
			return fapp, ok
		}
		applications = append(applications, fapp)
	}
	app := applications[len(applications)-1]
	for i := len(applications) - 2; i >= 0; i-- {
		napp := applications[i]
		span := span.NewSpan(napp.NodeSpan().Beg, app.NodeSpan().End)
		app = &ast.FuncApplication{
			Span:   &span,
			Callee: napp,
			Args:   []ast.Expr{app},
		}
	}
	return app, true
}

func (p *Parser) parseBlock() (*ast.Block, bool) {
	log.Println("Parsing block")
	beg := p.position()
	if p.match(token.Colon) == nil {
		return nil, true
	}
	if p.match(token.NewLine) == nil {
		p.error(beg, p.position(), "expected new line for a block")
		p.recoverWithTokens(token.NewLine)
		p.match(token.NewLine)
		return nil, false
	}
	indent, err := p.pushNextIndent()
	if err != nil {
		p.error(beg, p.position(), "expected block instructions to be indented")
		p.recover()
		return nil, false
	}
	defer p.popIndent(indent)
	parseStmt := func() (ast.Stmt, bool) {
		p.skipEmptyLines()
		log.Printf("%d running wrapped parse", indent)
		if t := p.matchIndent(indent); !t {
			log.Println("Indentation does not match")
			return nil, true
		}
		log.Println("Parse stmt for block")
		return p.parseStmt()
	}
	exprs := []ast.Stmt{}
	var e ast.Stmt = nil
	ok := true
	for e, ok = parseStmt(); e != nil && ok; e, ok = parseStmt() {
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
	fn, ok := p.parseSimpleExpr()
	if fn == nil || !ok {
		return nil, ok
	}
	arg, ok := p.parseSimpleExpr()
	if !ok {
		return nil, false
	}
	potentialTrailing := p.check(token.Do) != nil || p.check(token.Colon) != nil
	hasTrailingBlock := p.parseTrailingBlocks && potentialTrailing
	if arg == nil && !hasTrailingBlock {
		// not a function call, just normal expression
		return fn, true
	}
	args := []ast.Expr{}
	for arg != nil {
		args = append(args, arg)
		arg, ok = p.parseSimpleExpr()
		if !ok {
			return nil, false
		}
	}
	block, ok := p.parseTrailingLambda()
	if !ok {
		return nil, ok
	}
	span := span.NewSpan(beg, p.position())
	node := ast.FuncApplication{
		Span:   &span,
		Callee: fn,
		Args:   args,
		Block:  block,
	}
	return &node, true
}

func (p *Parser) parseTrailingLambda() (*ast.LambdaExpr, bool) {
	var block *ast.LambdaExpr = nil
	if p.parseTrailingBlocks && p.check(token.Colon) != nil {
		p.openScope()
		defer p.closeScope()
		lbeg := p.position()
		b, ok := p.parseBlock()
		if !ok {
			return nil, false
		}
		if b == nil {
			p.error(lbeg, p.position(), "expected anonymous block")
			p.recover()
		}
		span := span.NewSpan(lbeg, p.position())
		block = &ast.LambdaExpr{
			Span: &span,
			Args: []*ast.FuncDeclArg{},
			Body: b,
		}
	} else {
		l, ok := p.parseLambda()
		if !ok {
			return nil, false
		}
		if l != nil {
			block, ok = l.(*ast.LambdaExpr)
			if !ok {
				panic("unreachable")
			}
		}
	}
	return block, true
}

func (p *Parser) parseSimpleExpr() (ast.Expr, bool) {
	log.Println("Parse simple expression")
	beg := p.position()
	node, ok := p.parsePrimaryExpr()
	if node == nil || !ok {
		return node, ok
	}
loop:
	for {
		log.Println("Checking for postfix operator")
		t := p.curr
		switch t.Typ {
		case token.Exclamation:
			log.Println("It's nullary application")
			p.bump()
			span := span.NewSpan(beg, p.position())
			node = &ast.FuncApplication{
				Span:   &span,
				Callee: node,
				Args:   make([]ast.Expr, 0),
				Block:  nil,
			}
		case token.Access:
			p.bump()
			log.Println("It's access")
			id := p.parseIdentifier()
			if id == nil {
				p.error(beg, p.position(), "expected indentifier in access expression")
				return nil, !ok
			}
			span := span.NewSpan(beg, p.position())
			node = &ast.Access{
				Span:     &span,
				Lhs:      node,
				Property: *id,
			}
		default:
			log.Printf("No postfixes, token is %s\n", token.IdToString(t.Typ))
			break loop
		}
	}
	return node, ok
}

func (p *Parser) parsePrimaryExpr() (ast.Expr, bool) {
	log.Println("Parse primary expression")
	beg := p.position()
	tok := p.curr
	log.Printf("Token is %s\n", token.IdToString(tok.Typ))
	switch tok.Typ {
	case token.LParen:
		// todo: make parenthesis as well as tuple multiline
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
		p.tryLiftVar(node.Name)
		return &node, true
	case token.LBracket:
		p.bump()
		return p.parseRecordConst(beg)
	case token.LSquareParen:
		p.bump()
		return p.parseListConst(beg)
	case token.True:
		p.bump()
		var node ast.BoolConst
		node.Span = tok.Span
		node.Val = true
		return &node, true
	case token.False:
		p.bump()
		var node ast.BoolConst
		node.Span = tok.Span
		node.Val = false
		return &node, true
	case token.None:
		p.bump()
		var node ast.NoneConst
		node.Span = tok.Span
		return &node, true
	case token.Quote:
		p.bump()
		id := p.match(token.Identifier)
		if id == nil {
			p.error(beg, p.position(), "Only identifiers can be quoted")
			return nil, false
		}
		return &ast.Symbol{Span: id.Span, Val: id.Val}, true
	default:
		log.Println("Not a primary")
		return nil, true
	}
}

func (p *Parser) parseWhile() (ast.Stmt, bool) {
	beg := p.position()
	log.Printf("Parsing while, curr token is %s", token.IdToString(p.curr.Typ))
	if t := p.match(token.While); t == nil {
		return nil, true
	}
	p.disallowTrailingBlocks()
	cond, ok := p.parseExpr()
	p.allowTrailingBlocks()
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
	node := ast.WhileStmt{
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
	p.disallowTrailingBlocks()
	cond, ok := p.parseExpr()
	p.allowTrailingBlocks()
	if !ok {
		return nil, false
	}
	if cond == nil {
		p.error(beg, p.position(), "if expects an expression as its condition")
		p.recoverWithTokens(token.Colon)
	}
	body, ok := p.parseBlock()
	if !ok {
		return nil, false
	}
	if body == nil {
		p.error(beg, p.position(), "if expects a block as its body")
		p.recover()
		return nil, true
	}
	elseb, ok := p.parseElse()
	if !ok {
		p.recover()
		return nil, true
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
	if p.match(token.Else) == nil {
		if !p.checkIndent(p.currentIndent()) {
			return nil, true
		}
		if p.peek().Typ != token.Else {
			return nil, true
		}
		p.bump()
		p.bump()
	}
	if p.curr.Typ == token.If {
		return p.parseIf()
	}
	return p.parseBlock()
}

func (p *Parser) parseHandle() (ast.Expr, bool) {
	log.Println("Parsing handle")
	beg := p.position()
	if p.match(token.Handle) == nil {
		log.Println("not a handle")
		return nil, true
	}
	p.openScope()
	body, ok := p.parseBlock()
	p.closeScope()
	if !ok {
		return nil, false
	}
	if body == nil {
		p.error(beg, p.position(), "handle expects a block as its body")
		return nil, true
	}
	ww := make([]*ast.WithClause, 0, 1)
	cw, ok := p.parseWith()
	for cw != nil {
		if !ok {
			p.error(beg, p.position(), "handle expects at least one with")
			return nil, true
		}
		ww = append(ww, cw)
		cw, ok = p.parseWith()
	}
	span := span.NewSpan(beg, p.position())
	return &ast.Handle{
		Span: &span,
		Body: body,
		Arms: ww,
	}, true
}

func (p *Parser) parseWith() (*ast.WithClause, bool) {
	beg := p.position()
	if p.match(token.With) == nil {
		if !p.checkIndent(p.currentIndent()) {
			return nil, true
		}
		if p.peek().Typ != token.With {
			return nil, true
		}
		p.bump()
		p.bump()
	}
	effect, ok := p.parsePath()
	if effect == nil || !ok {
		p.error(beg, p.position(), "Expected effect to handle in with clause")
		p.recoverWithTokens(token.Colon)
	}
	p.openScope()
	argid := p.parseIdentifier()
	arg := &ast.FuncDeclArg{}
	if argid == nil {
		p.error(beg, p.position(), "expected effect's value name")
		p.recoverWithTokens(token.Colon)
	} else {
		arg.Span = argid.Span
		arg.Name = argid.Name
		p.scope.InsertFuncArg(arg)
	}
	defer p.closeScope()
	p.scope.InsertFuncArg(arg)
	var cont *ast.FuncDeclArg = nil
	if p.match(token.Arrow) != nil {
		contid := p.parseIdentifier()
		if contid == nil {
			p.error(beg, p.position(), "Expected name for the continuation")
			p.recoverWithTokens(token.Colon)
		} else {
			cont = &ast.FuncDeclArg{
				Span: contid.Span,
				Name: contid.Name,
			}
			p.scope.InsertFuncArg(cont)
		}
	}
	b, ok := p.parseBlock()
	if b == nil && ok {
		p.error(beg, p.position(), "Expected block as with stmt's body")
		p.recoverWithTokens(token.NewLine)
		p.match(token.NewLine)
		return nil, true
	}
	span := span.NewSpan(beg, p.position())
	return &ast.WithClause{
		Span:         &span,
		Effect:       effect,
		Arg:          arg,
		Continuation: cont,
		Body:         b,
	}, ok
}

func (p *Parser) parsePath() (ast.Expr, bool) {
	log.Println("Parse path expression")
	beg := p.position()
	id := p.parseIdentifier()
	if id == nil {
		return nil, true
	}
	var node ast.Expr = id
loop:
	for {
		log.Println("Checking for postfix dot operator")
		t := p.curr
		switch t.Typ {
		case token.Access:
			p.bump()
			log.Println("It's access")
			id := p.parseIdentifier()
			if id == nil {
				p.error(beg, p.position(), "expected indentifier in access expression")
				return nil, false
			}
			span := span.NewSpan(beg, p.position())
			node = &ast.Access{
				Span:     &span,
				Lhs:      node,
				Property: *id,
			}
		default:
			log.Printf("No postfixes, token is %s\n", token.IdToString(t.Typ))
			break loop
		}
	}
	return node, true
}

func (p *Parser) parseLambda() (ast.Expr, bool) {
	log.Println("Parsing lambda")
	beg := p.position()
	p.openScope()
	defer p.closeScope()
	if p.match(token.Do) == nil {
		log.Println("Not a lambda")
		return nil, true
	}
	args := []*ast.FuncDeclArg{}
	if t := p.match(token.Pipe); t != nil {
		log.Println("Parsing lambda arguments")
		for pt := p.match(token.Pipe); pt == nil; pt = p.match(token.Pipe) {
			arg := p.parseIdentifier()
			if arg == nil {
				p.error(beg, p.position(), "Lambda argument has to be an identifier")
				p.recoverWithTokens(token.Pipe, token.Colon, token.Arrow)
				p.match(token.Pipe)
				break
			}
			a := &ast.FuncDeclArg{
				Span: arg.Span,
				Name: arg.Name,
			}
			log.Printf("Parsed parameter %s", a.Name)
			args = append(args, a)
			p.scope.InsertFuncArg(a)
		}
	} else {
		log.Println("Parsing lambda arguments witout pipe")
		for {
			arg := p.parseIdentifier()
			if arg == nil {
				if p.check(token.Colon) != nil || p.check(token.Arrow) != nil {
					break
				} else {
					p.error(beg, p.position(), "Lambda argument has to be an identifier")
					p.recoverWithTokens(token.Pipe, token.Colon, token.Arrow)
					p.match(token.Pipe)
					break
				}
			}
			a := &ast.FuncDeclArg{
				Span: arg.Span,
				Name: arg.Name,
			}
			log.Printf("Parsed parameter %s", a.Name)
			args = append(args, a)
			p.scope.InsertFuncArg(a)
		}
	}
	log.Println("Parsed lambda arguments")
	var body ast.Expr
	if p.parseTrailingBlocks && p.check(token.Colon) != nil {
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
	} else if p.match(token.Arrow) != nil {
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

func (p *Parser) parseLocalFnDecl() (*ast.LambdaExpr, bool) {
	log.Println("Parse local fn decl")
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
	p.openScope()
	p.scope.Insert(name.Name)
	defer p.closeScope()

	args := []*ast.FuncDeclArg{}
	for arg := p.parseIdentifier(); arg != nil; arg = p.parseIdentifier() {
		farg := &ast.FuncDeclArg{Span: arg.Span, Name: arg.Name}
		args = append(args, farg)
		p.scope.InsertFuncArg(farg)
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
	fn := ast.LambdaExpr{
		Span: &span,
		Name: name.Name,
		Args: args,
		Body: fbody,
	}
	return &fn, true
}

func (p *Parser) parseValDecl() (ast.Stmt, bool) {
	log.Println("Parsing local val decl")
	beg := p.position()
	if t := p.match(token.Let); t == nil {
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
	p.match(token.NewLine)
	span := span.NewSpan(beg, p.position())
	node := ast.ValDecl{
		Span: &span,
		Name: name.Val,
		Rhs:  expr,
	}
	p.scope.InsertVal(&node)
	return &node, ok
}

func (p *Parser) parseLocalEffectDecl() (*ast.LocalEffect, bool) {
	beg := p.position()
	if p.match(token.Effect) == nil {
		return nil, true
	}
	name := p.parseIdentifier()
	if name == nil {
		p.error(beg, p.position(), "Expected idnetifier as an effect name")
		return nil, false
	}
	span := span.NewSpan(beg, p.position())
	if p.scope.IsGlobal() {
		panic("ICE: expected local scope")
	}
	return &ast.LocalEffect{
		Name: name.Name,
		Span: &span,
	}, true
}

func (p *Parser) parseReturn() (ast.Stmt, bool) {
	beg := p.position()
	if p.match(token.Return) == nil {
		return nil, true
	}
	v, ok := p.parseExpr()
	span := span.NewSpan(beg, p.position())
	if v == nil {
		v = &ast.NoneConst{Span: &span}
	}
	ret := &ast.Return{
		Span: &span,
		Val:  v,
	}
	return ret, ok
}

func (p *Parser) parseResume() (ast.Expr, bool) {
	beg := p.position()
	if p.match(token.Resume) == nil {
		return nil, true
	}
	cont, ok := p.parseSimpleExpr()
	if !ok {
		p.error(beg, p.position(), "Resume expects at least a continuation to run with")
		p.recoverWithTokens(token.NewLine)
		return nil, false
	}
	arg, ok := p.parseSimpleExpr()
	if !ok {
		return nil, false
	}
	span := span.NewSpan(beg, p.position())
	return &ast.Resume{
		Span: &span,
		Cont: cont,
		Arg:  arg,
	}, true
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

func (p *Parser) parseRecordConst(beg span.Position) (*ast.RecordConst, bool) {
	log.Println("Parsing record literal")
	vals := []ast.RecordField{}
	continuationIndent := -1
	tryParseIndent := func() bool {
		log.Println("Trying to pass possible indent")
		if p.match(token.NewLine) == nil {
			log.Println("No new line, means nothing to do")
			return true
		}
		p.skipEmptyLines()
		if continuationIndent == -1 {
			i, err := p.pushNextIndent()
			if err != nil {
				p.error(beg, p.position(), "expected indentation as record continuation")
			}
			log.Printf("Got new line, indentation for this expression is %d\n", i)
			continuationIndent = i
		}
		v := p.matchIndent(continuationIndent)
		log.Printf("Matching intentetion %d, matched=%v\n", continuationIndent, v)
		return v
	}
	for {
		if !tryParseIndent() {
			break
		}
		key := p.parseIdentifier()
		if key != nil {
			log.Printf("Parsing expr for key %s\n", key)
			if p.match(token.Colon) == nil {
				// maybe same name variable sugar?
				if p.check(token.Comma) != nil || p.check(token.RBracket) != nil {
					vals = append(vals, ast.RecordField{Key: key.Name, Val: key})
				} else {
					p.error(beg, p.position(), "missing colon \":\" after key in record literal")
					p.recoverWithTokens(token.NewLine, token.RBracket)
					continue
				}
			} else {
				val, ok := p.parseExpr()
				if val == nil || !ok {
					p.error(beg, p.position(), "record literal expects an expression as its values")
					log.Println("Record literal could not parse expression, recovering.")
					p.recoverWithTokens(token.NewLine, token.RBracket)
					continue
				}
				vals = append(vals, ast.RecordField{Key: key.Name, Val: val})
			}
		} else {
			// function syntax sugar
			f, ok := p.parseLocalFnDecl()
			if !ok {
				log.Println("Error while parsing fn declaration sugar in record")
				p.recoverWithTokens(token.Comma, token.RBracket)
				continue
			}
			if f == nil {
				break
			}
			vals = append(vals, ast.RecordField{Key: f.Name, Val: f})
		}
		if p.match(token.Comma) == nil {
			if p.check(token.NewLine) != nil {
				p.error(beg, p.position(), "missing comma (,) in record literal")
				continue
			}
			break
		}
	}
	if continuationIndent != -1 {
		p.popIndent(continuationIndent)
	}
	if p.match(token.RBracket) == nil {
		if p.checkIndent(p.currentIndent()) {
			if p.peek().Typ == token.RBracket {
				p.bump()
				p.bump()
			} else {
				p.error(beg, p.position(), "missing closing bracket in record literal")
			}
		}
	}
	span := span.NewSpan(beg, p.position())
	rec := ast.RecordConst{
		Span:   &span,
		Fields: vals,
	}
	return &rec, true
}

func (p *Parser) parseListConst(beg span.Position) (*ast.ListConst, bool) {
	log.Println("Parsing list literal")
	vals := []ast.Expr{}
	continuationIndent := -1
	tryParseIndent := func() bool {
		log.Println("Trying to pass possible indent")
		if p.match(token.NewLine) == nil {
			log.Println("No new line, means nothing to do")
			return true
		}
		p.skipEmptyLines()
		if continuationIndent == -1 {
			i, err := p.pushNextIndent()
			if err != nil {
				p.error(beg, p.position(), "expected indentation as a list continuation.\nMaybe you meant empty list? \"[]\"")
			}
			log.Printf("Got new line, indentation for this expression is %d\n", i)
			continuationIndent = i
		}
		v := p.matchIndent(continuationIndent)
		log.Printf("Matching intentetion %d, matched=%v\n", continuationIndent, v)
		return v
	}
	for {
		if !tryParseIndent() {
			break
		}
		log.Println("Parsing new expression for list")
		e, ok := p.parseExpr()
		if e == nil {
			if !ok {
				p.error(beg, p.position(), "Expected expression in list literal")
				p.recoverWithTokens(token.NewLine, token.RSquareParen)
				continue
			}
			break
		}
		vals = append(vals, e)
		if p.match(token.Comma) == nil {
			break
		}
	}
	if continuationIndent != -1 {
		p.popIndent(continuationIndent)
	}
	if p.match(token.RSquareParen) == nil {
		if p.checkIndent(p.currentIndent()) {
			if p.peek().Typ == token.RSquareParen {
				p.bump()
				p.bump()
			} else {
				p.error(beg, p.position(), "Expected ] to close a list literal")
			}
		}
	}
	span := span.NewSpan(beg, p.position())
	list := &ast.ListConst{
		Span: &span,
		Vals: vals,
	}
	return list, true
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
				p.match(token.Comma)
				continue
			}
			break
		}
		log.Printf("Parsed tuple val %s\n", e.String())
		vals = append(vals, e)
		if p.match(token.Comma) == nil {
			break
		}
	}
	log.Println("Endend tuple parsing")
	if p.match(token.RParen) == nil {
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

func (p *Parser) peek() *token.Token {
	return &p.l.peek
}

func (p *Parser) check(typ token.Id) *token.Token {
	if p.curr.Typ == typ {
		tok := p.curr
		return &tok
	}
	return nil
}

// todo: Maybe we can do something here?
// We always skip empty lines and if we have new indentation pushed
// we just skip any indentation that matches it?
func (p *Parser) match(typ token.Id) *token.Token {
	ist := p.check(typ)
	if ist != nil {
		p.bump()
	}
	return ist
}

func (p *Parser) checkIndent(n int) bool {
	return p.checkIndentWith(func(v int) bool { return v == n })
}

func (p *Parser) checkIndentWith(pred func(int) bool) bool {
	t := p.l.Current()
	if t.Typ != token.Indent {
		log.Printf("token is not an indentation %s\n", token.IdToString(t.Typ))
		return false
	}
	val, err := strconv.Atoi(t.Val)
	if err != nil {
		log.Panicf("could not parse indentations value, %s", t.Val)
	}
	if !pred(val) {
		log.Printf("indent not matching the predicate - got: %d", val)
		return false
	}
	return true
}

func (p *Parser) matchIndent(n int) bool {
	return p.matchIndentWith(func(v int) bool { return v == n })
}

func (p *Parser) matchIndentWith(pred func(int) bool) bool {
	isind := p.checkIndentWith(pred)
	if isind {
		p.bump()
	}
	return isind
}

func (p *Parser) bump() {
	p.curr = p.l.Next()
}

func (p *Parser) currentIndent() int {
	if len(p.indents) == 0 {
		return 0
	}
	return p.indents[len(p.indents)-1]
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

// recoverWithTokens eats tokens until it sees any of
// the passed control tokens. It does not eat the control token
// unless it hit end of top level item.
func (p *Parser) recoverWithTokens(rtt ...token.Id) {
	log.Println("In recover")
	defer p.l.UnsetMode(skipErrorReporting)
	p.l.SetMode(skipErrorReporting)
	var t token.Token
	eat := false
	for t = p.l.curr; !p.eof() && !isRecoveryToken(t, rtt); t = p.l.Next() {
		log.Println("Recovering")
		if t.Typ == token.NewLine && p.l.Peek().Typ != token.Indent {
			eat = true
			break
		}
	}
	if eat {
		p.bump()
	}
	log.Println("Recovered")
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
	return p.l.curr.Span.Beg
}

func (p *Parser) disallowTrailingBlocks() {
	p.parseTrailingBlocks = false
}

func (p *Parser) allowTrailingBlocks() {
	p.parseTrailingBlocks = true
}

func (p *Parser) closeScope() {
	if p.scope.IsGlobal() {
		panic("ICE: cannot pop a global scope")
	}
	p.scope = p.scope.parent
}

func (p *Parser) openScope() {
	p.scope = p.scope.Derive()
}

func (p *Parser) tryLiftVar(name string) {
	rs, si := p.scope.RelativeScope(name)
	if rs == Outer {
		si.Lift()
	}
}

func (p *Parser) skipEmptyLines() {
	for p.skipEmptyLine() {
	}
}

func (p *Parser) skipEmptyLine() bool {
	if p.check(token.Indent) != nil && p.peek().Typ == token.NewLine {
		p.bump()
		p.bump()
		return true
	}
	nl := p.check(token.NewLine)
	if nl != nil && nl.Span.Beg.Column == 0 {
		p.bump()
		return true
	}
	return false
}
