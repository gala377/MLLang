package codegen

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
	"github.com/gala377/MLLang/syntax"
	"github.com/gala377/MLLang/syntax/ast"
	"github.com/gala377/MLLang/syntax/span"
)

type CompilationError struct {
	Location *span.Span
	Message  string
}

func (c CompilationError) Error() string {
	return c.Message
}

func (c CompilationError) SourceLoc() span.Span {
	return *c.Location
}

type Emitter struct {
	result   *data.Code
	line     int
	interner *Interner
	errors   []CompilationError
	scope    *syntax.Scope
}

func NewEmitter(i *Interner) *Emitter {
	c := data.NewCode()
	e := Emitter{
		result:   &c,
		line:     0,
		interner: i,
		errors:   make([]CompilationError, 0),
		scope:    syntax.NewScope(nil),
	}
	return &e
}

func (e *Emitter) Compile(nn []ast.Node) (*data.Code, []CompilationError) {
	for _, n := range nn {
		e.emitNode(n)
	}
	if len(e.errors) > 0 {
		return nil, e.errors
	}
	return e.result, nil
}

func (e *Emitter) error(loc *span.Span, msg string) {
	e.errors = append(e.errors, CompilationError{
		Location: loc,
		Message:  msg,
	})
}

func (e *Emitter) Interner() *Interner {
	return e.interner
}

func (e *Emitter) emitNode(n ast.Node) {
	e.line = int(n.NodeSpan().Beg.Line)
	if v, ok := n.(ast.Stmt); ok {
		e.emitStmt(v)
		return
	}
	if v, ok := n.(ast.Decl); ok {
		e.emitDeclaration(v)
		return
	}
	e.error(n.NodeSpan(), "Compiling this node is not supported")
}

func (e *Emitter) emitByte(b byte) {
	e.result.WriteByte(b, e.line)
}
func (e *Emitter) emitBytes(bb ...byte) {
	for _, b := range bb {
		e.emitByte(b)
	}
}

func (e *Emitter) emitConstant(v data.Value) {
	index := e.result.AddConstant(v)
	if index > 255 {
		if index > math.MaxUint16 {
			// rare panic but that is the best we can do honestly
			panic("More constants that uint16 can hold. That is not supported.")
		}
		e.emitByte(isa.Constant2)
		args := []byte{0, 0}
		binary.BigEndian.PutUint16(args, uint16(index))
		e.emitBytes(args...)
		return
	}
	e.emitBytes(isa.Constant, byte(index))
}

func (e *Emitter) emitExpr(node ast.Expr) {
	e.line = int(node.NodeSpan().Beg.Line)
	switch v := node.(type) {
	case *ast.IfExpr:
		e.emitIf(v)
	case *ast.IntConst:
		e.emitIntConst(v)
	case *ast.BoolConst:
		e.emitBoolConst(v)
	case *ast.StringConst:
		e.emitStringConst(v)
	case *ast.Block:
		e.emitBlock(v)
	case *ast.Identifier:
		if e.scope.LookupLocal(v.Name) {
			e.emitLocalLookup(v)
		} else {
			e.emitGlobalLookup(v)
		}
	case *ast.FuncApplication:
		e.emitApplication(v)
	case *ast.LambdaExpr:
		e.emitLambda(v)
	default:
		log.Printf("Node is %v", node)
		e.error(node.NodeSpan(), "Node cannot be emitted. Not supported")
	}
}

func (e *Emitter) emitStmt(node ast.Stmt) {
	e.line = int(node.NodeSpan().Beg.Line)
	switch v := node.(type) {
	case *ast.StmtExpr:
		log.Printf("Got StmtExpression")
		e.emitUnboundExpr(v.Expr)
	case *ast.ValDecl:
		e.emitVariableDecl(v)
	default:
		log.Printf("Stmt node is %v", node)
		e.error(node.NodeSpan(), "Stmt node cannot be emitted. Not supported")
	}
}

func (e *Emitter) emitUnboundExpr(node ast.Expr) {
	e.emitExpr(node)
	e.emitByte(isa.Pop)
}

func (e *Emitter) emitIf(node *ast.IfExpr) {
	e.emitExpr(node.Cond)
	ifpos := e.emitJumpIfFalse()
	e.emitBlock(node.IfBranch)
	if node.ElseBranch != nil {
		skipElse := e.emitJump()
		off := e.result.Len() - ifpos
		e.patchJump(ifpos, off)
		e.emitExpr(node.ElseBranch)
		off = e.result.Len() - skipElse
		e.patchJump(skipElse, off)
		return
	}
	skipNone := e.emitJump()
	off := e.result.Len() - ifpos
	e.patchJump(ifpos, off)
	e.emitNone()
	off = e.result.Len() - skipNone
	e.patchJump(skipNone, off)
}

func (e *Emitter) emitBlock(node *ast.Block) {
	for i, instr := range node.Instr {
		if i == (len(node.Instr) - 1) {
			// last in a block
			switch v := instr.(type) {
			case *ast.StmtExpr:
				e.emitExpr(v.Expr)
			default:
				e.emitStmt(v)
				e.emitByte(isa.PushNone)
			}
		} else {
			e.emitStmt(instr)
		}
	}
}

func (e *Emitter) emitNone() {
	e.emitConstant(data.None)
}

func (e *Emitter) emitIntConst(node *ast.IntConst) {
	v := data.NewInt(node.Val)
	e.emitConstant(v)
}

func (e *Emitter) emitBoolConst(node *ast.BoolConst) {
	v := data.NewBool(node.Val)
	e.emitConstant(v)
}

func (e *Emitter) emitStringConst(node *ast.StringConst) {
	v := data.NewString(node.Val)
	e.emitConstant(v)
}

func (e *Emitter) emitJumpIfFalse() int {
	e.emitBytes(isa.JumpIfFalse, 0, 0)
	return e.result.Len() - 3
}

func (e *Emitter) emitJump() int {
	e.emitBytes(isa.Jump, 0, 0)
	return e.result.Len() - 3
}

func (e *Emitter) patchJump(i int, offset int) {
	if offset > math.MaxUint16 {
		panic("Trying to jump more than uint16. Not supported")
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(offset))
	copy(e.result.Instrs[i+1:], args)
}

func (e *Emitter) emitGlobalLookup(node *ast.Identifier) {
	e.emitLookup(isa.DynLookup, node)
}

func (e *Emitter) emitLocalLookup(node *ast.Identifier) {
	e.emitLookup(isa.LocalLookup, node)
}

func (e *Emitter) emitLookup(kind isa.Op, node *ast.Identifier) {
	s := e.interner.Intern(node.Name)
	v := data.NewSymbol(s)
	index := e.result.AddConstant(v)
	if index > math.MaxUint16 {
		e.error(node.NodeSpan(), "More constants that uint16 can hold. That is not supported.")
		return
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(kind)
	e.emitBytes(args...)
}

func (e *Emitter) emitApplication(node *ast.FuncApplication) {
	e.emitExpr(node.Callee)
	for _, a := range node.Args {
		e.emitExpr(a)
	}
	argc := len(node.Args)
	if node.Block != nil {
		argc += 1
		e.emitLambda(node.Block)
	}
	if argc > 255 {
		e.error(node.NodeSpan(), "Function application with more than 255 arguments is not supported")
		e.emitBytes(isa.Call, byte(255))
		return
	}
	as := byte(argc)
	e.emitBytes(isa.Call, as)
}

func (e *Emitter) emitDeclaration(node ast.Decl) {
	switch v := node.(type) {
	case *ast.GlobalValDecl:
		e.emitGlobalVariableDecl(v)
	case *ast.FuncDecl:
		e.emitFuncDeclaration(v)
	default:
		panic("unreachable")
	}
}

func (e *Emitter) emitGlobalVariableDecl(node *ast.GlobalValDecl) {
	if !e.scope.IsGlobal() {
		panic("ICE: trying to emit global variable declation not in global scope")
	}
	if e.scope.Lookup(node.Name) {
		e.error(node.Span, fmt.Sprintf("redeclaration of name %s", node.Name))
		return
	}
	e.scope.Insert(node.Name)
	e.emitExpr(node.Rhs)
	s := e.interner.Intern(node.Name)
	v := data.NewSymbol(s)
	index := e.result.AddConstant(v)
	if index > math.MaxUint16 {
		e.error(node.NodeSpan(), "More constants that uint16 can hold. That is not supported.")
		return
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(isa.DefGlobal)
	e.emitBytes(args...)
}

func (e *Emitter) emitFuncDeclaration(node *ast.FuncDecl) {
	if !e.scope.IsGlobal() {
		panic("ICE: trying to define function in local scope")
	}
	if e.scope.Lookup(node.Name) {
		e.error(node.NodeSpan(), fmt.Sprintf("redeclaration of the name %s", node.Name))
	}
	e.scope.Insert(node.Name)
	fname := data.NewSymbol(e.interner.Intern(node.Name))
	// emit function body
	fe := NewEmitter(e.interner)
	fe.scope = e.scope.Derive()
	fargs := make([]data.Symbol, 0, len(node.Args))
	for _, arg := range node.Args {
		fe.scope.Insert(arg.Name)
		s := e.interner.Intern(arg.Name)
		fargs = append(fargs, data.NewSymbol(s))
	}
	fe.emitExpr(node.Body)
	// todo: implicit return might not always be needed but then
	// we will never get there if there is an explicit one
	fe.emitByte(isa.Return)
	e.errors = append(e.errors, fe.errors...)
	code := fe.result
	l := data.NewFunction(fname, fargs, code)
	index := e.result.AddConstant(l)
	if index > math.MaxUint16 {
		e.error(node.NodeSpan(), "More constants that uint16 can hold. That is not supported.")
		return
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(isa.Lambda)
	e.emitBytes(args...)
	// assign to global variable
	index = e.result.AddConstant(fname)
	if index > math.MaxUint16 {
		e.error(node.NodeSpan(), "More constants that uint16 can hold. That is not supported.")
		return
	}
	args = []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(isa.DefGlobal)
	e.emitBytes(args...)
}

func (e *Emitter) emitVariableDecl(node *ast.ValDecl) {
	if e.scope.IsGlobal() {
		panic("ICE: trying to emit local val declaration in global scope")
	}
	if e.scope.LookupLocal(node.Name) {
		e.error(node.NodeSpan(), fmt.Sprintf("redeclaration of local name %s", node.Name))
		return
	}
	e.scope.Insert(node.Name)
	e.emitExpr(node.Rhs)
	s := e.interner.Intern(node.Name)
	v := data.NewSymbol(s)
	index := e.result.AddConstant(v)
	if index > math.MaxUint16 {
		e.error(node.NodeSpan(), "More constants that uint16 can hold. That is not supported.")
		return
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(isa.DefLocal)
	e.emitBytes(args...)
}

func (e *Emitter) emitLambda(node *ast.LambdaExpr) {
	le := NewEmitter(e.interner)
	le.scope = e.scope.Derive()
	fargs := make([]data.Symbol, 0, len(node.Args))
	for _, arg := range node.Args {
		le.scope.Insert(arg.Name)
		s := e.interner.Intern(arg.Name)
		fargs = append(fargs, data.NewSymbol(s))
	}
	le.emitExpr(node.Body)
	// todo: implicit return might not always be needed but then
	// we will never get there if there is an explicit one
	le.emitByte(isa.Return)
	e.errors = append(e.errors, le.errors...)
	code := le.result
	l := data.NewLambda(nil, fargs, code)
	index := e.result.AddConstant(l)
	if index > math.MaxUint16 {
		e.error(node.NodeSpan(), "More constants that uint16 can hold. That is not supported.")
		return
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(isa.Lambda)
	e.emitBytes(args...)
}
