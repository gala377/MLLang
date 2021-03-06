package codegen

import (
	"encoding/binary"
	"errors"
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
	result         *data.Code
	line           int
	interner       *Interner
	errors         []CompilationError
	scope          *syntax.Scope
	path           string
	inTailPosition bool
	counter        int
}

func NewEmitter(path string, i *Interner) *Emitter {
	c := data.NewCode()
	c.Path = path
	e := Emitter{
		result:   &c,
		line:     0,
		interner: i,
		errors:   make([]CompilationError, 0),
		// todo share scope from parser
		scope:          syntax.NewScope(nil),
		path:           path,
		inTailPosition: false,
		counter:        0,
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
		e.emitIf(v, false)
	case *ast.IntConst:
		e.emitIntConst(v)
	case *ast.FloatConst:
		e.emitFloatConst(v)
	case *ast.BoolConst:
		e.emitBoolConst(v)
	case *ast.StringConst:
		e.emitStringConst(v)
	case *ast.NoneConst:
		e.emitNone()
	case *ast.Block:
		e.emitBlock(v, false)
	case *ast.Identifier:
		if si := e.scope.LookupLocal(v.Name); si != nil {
			e.emitLocalLookup(v, si)
		} else {
			e.emitGlobalLookup(v)
		}
	case *ast.FuncApplication:
		e.emitApplication(v, false)
	case *ast.LambdaExpr:
		e.emitLambda(v)
	case *ast.ListConst:
		e.emitSequence(isa.MakeList, v)
	case *ast.TupleConst:
		e.emitSequence(isa.MakeTuple, v)
	case *ast.RecordConst:
		e.emitRecord(v)
	case *ast.Access:
		e.emitAccess(v)
	case *ast.Symbol:
		e.emitSymbol(v.Val)
	case *ast.Handle:
		e.emitHandler(v, false)
	case *ast.LocalEffect:
		e.emitLocalEffect(v)
	case *ast.Resume:
		e.emitResume(v, false)
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
	case *ast.Assignment:
		e.emitAssignment(v)
	case *ast.WhileStmt:
		e.emitWhile(v)
	case *ast.Return:
		e.emitReturn(v)
	default:
		log.Printf("Stmt node is %v", node)
		e.error(node.NodeSpan(), "Stmt node cannot be emitted. Not supported")
	}
}

func (e *Emitter) emitUnboundExpr(node ast.Expr) {
	e.emitExpr(node)
	e.emitByte(isa.Pop)
}

func (e *Emitter) emitNone() {
	e.emitByte(isa.PushNone)
}

func (e *Emitter) emitIntConst(node *ast.IntConst) {
	v := data.NewInt(node.Val)
	e.emitConstant(v)
}

func (e *Emitter) emitFloatConst(node *ast.FloatConst) {
	v := data.NewFloat(node.Val)
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

func (e *Emitter) emitSymbol(val string) {
	v := data.NewSymbol(e.interner.Intern(val))
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

func (e *Emitter) emitJumpBack() int {
	e.emitBytes(isa.JumpBack, 0, 0)
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
	e.emitLookup(isa.LoadDyn, node)
}

func (e *Emitter) emitLocalLookup(node *ast.Identifier, si syntax.ScopeInfo) {
	if si.IsLifted() {
		e.emitLookup(isa.LoadDeref, node)
	} else {
		e.emitLookup(isa.LoadLocal, node)
	}
}
func (e *Emitter) emitLookup(kind isa.Op, node *ast.Identifier) {
	index, err := e.addSymbol(node.Name)
	if err != nil {
		e.error(node.NodeSpan(), err.Error())
		return
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(kind)
	e.emitBytes(args...)
}

func (e *Emitter) emitAssignment(node *ast.Assignment) {
	var index int
	var err error
	instr := isa.StoreDyn
	switch loc := node.LValue.(type) {
	case *ast.Identifier:
		index, err = e.addSymbol(loc.Name)
		if err != nil {
			e.error(node.NodeSpan(), err.Error())
			return
		}
		if si := e.scope.LookupLocal(loc.Name); si != nil {
			instr = isa.StoreLocal
			if si.IsLifted() {
				instr = isa.StoreDeref
			}
		}
	case *ast.Access:
		e.emitExpr(loc.Lhs)
		index, err = e.addSymbol(loc.Property.Name)
		if err != nil {
			e.error(node.NodeSpan(), err.Error())
		}
		instr = isa.SetField
	default:
		e.error(node.NodeSpan(), "values can only be assigned to names or properties")
		return
	}
	e.emitExpr(node.RValue)
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.line = int(node.Beg.Line)
	e.emitByte(instr)
	e.emitBytes(args...)
}

func (e *Emitter) emitApplication(node *ast.FuncApplication, tailpos bool) {
	e.emitExpr(node.Callee)
	call0 := isa.Call0
	call1 := isa.Call1
	if tailpos {
		call0 = isa.TailCall0
		call1 = isa.TailCall1
	}
	if len(node.Args) == 0 && node.Block == nil {
		e.line = int(node.Beg.Line)
		e.emitByte(call0)
		return
	}
	for i, a := range node.Args {
		e.emitExpr(a)
		e.line = int(node.Beg.Line)
		if i == len(node.Args)-1 && node.Block == nil {
			e.emitByte(call1)
		} else {
			e.emitByte(isa.Call1)
		}
	}
	if node.Block != nil {
		e.emitLambda(node.Block)
		e.line = int(node.Beg.Line)
		e.emitByte(call1)
	}
}

func (e *Emitter) emitDeclaration(node ast.Decl) {
	switch v := node.(type) {
	case *ast.GlobalValDecl:
		e.emitGlobalVariableDecl(v)
	case *ast.FuncDecl:
		e.emitFuncDeclaration(v)
	case *ast.EffectDecl:
		e.emitGlobalEffectDecl(v)
	default:
		panic("unreachable")
	}
}

func (e *Emitter) emitGlobalVariableDecl(node *ast.GlobalValDecl) {
	if !e.scope.IsGlobal() {
		panic("ICE: trying to emit global variable declation not in global scope")
	}
	if e.scope.Lookup(node.Name) != nil {
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
	e.line = int(node.Beg.Line)
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(isa.DefGlobal)
	e.emitBytes(args...)
}

func (e *Emitter) emitGlobalEffectDecl(node *ast.EffectDecl) {
	if !e.scope.IsGlobal() {
		panic("ICE: trying to emit global effect declration not in global scope")
	}
	if e.scope.Lookup(node.Name) != nil {
		e.error(node.Span, fmt.Sprintf("redeclaration of name %s", node.Name))
		return
	}
	e.scope.Insert(node.Name)
	iname := e.interner.Intern(node.Name)
	s := data.NewSymbol(iname)
	e.emitConstant(data.NewType(s))
	index, err := e.addSymbol(node.Name)
	if err != nil {
		e.error(node.Span, fmt.Sprintf("Cannot emit effect name %s", err))
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
	if e.scope.Lookup(node.Name) != nil {
		e.error(node.NodeSpan(), fmt.Sprintf("redeclaration of the name %s", node.Name))
	}
	e.scope.Insert(node.Name)
	fname := data.NewSymbol(e.interner.Intern(node.Name))
	// emit function body
	fe := NewEmitter(e.path, e.interner)
	fe.scope = e.scope.Derive()
	fargs := make([]data.Symbol, 0, len(node.Args))
	for _, arg := range node.Args {
		fe.scope.InsertFuncArg(arg)
		s := e.interner.Intern(arg.Name)
		fargs = append(fargs, data.NewSymbol(s))
	}
	fe.emitLiftingForFuncArgs(node.Args, fargs)
	fe.emitExprInTailPos(node.Body)
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
	e.line = int(node.Beg.Line)
	e.emitByte(isa.Closure)
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

func (e *Emitter) emitLiftingForFuncArgs(args []*ast.FuncDeclArg, fargs []data.Symbol) {
	for i, arg := range args {
		if arg.Lift {
			id := ast.Identifier{Name: arg.Name, Span: arg.Span}
			index := e.result.AddConstant(fargs[i])
			if index > math.MaxUint16 {
				e.error(arg.Span, "More constants that uint16 can hold. That is not supported.")
				return
			}
			args := []byte{0, 0}
			binary.BigEndian.PutUint16(args, uint16(index))
			e.emitLookup(isa.LoadLocal, &id)
			e.emitByte(isa.MakeCell)
			e.emitByte(isa.StoreLocal)
			e.emitBytes(args...)
		}
	}
}

func (e *Emitter) emitVariableDecl(node *ast.ValDecl) {
	if e.scope.IsGlobal() {
		panic("ICE: trying to emit local val declaration in global scope")
	}
	if e.scope.LookupLocal(node.Name) != nil {
		e.error(node.NodeSpan(), fmt.Sprintf("redeclaration of local name %s", node.Name))
		return
	}
	e.emitExpr(node.Rhs)
	e.line = int(node.Beg.Line)
	index, err := e.addSymbol(node.Name)
	if err != nil {
		e.error(node.NodeSpan(), err.Error())
		return
	}
	e.scope.InsertVal(node)
	if node.Lift {
		e.emitByte(isa.MakeCell)
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(isa.DefLocal)
	e.emitBytes(args...)
}

func (e *Emitter) emitLambda(node *ast.LambdaExpr) {
	le := NewEmitter(e.path, e.interner)
	le.scope = e.scope.Derive()
	name := data.NewSymbol(nil)
	if node.Name != "" {
		// todo: probably will need lifting information
		le.scope.Insert(node.Name)
		name = data.NewSymbol(e.interner.Intern(node.Name))
	}
	fargs := make([]data.Symbol, 0, len(node.Args))
	for _, arg := range node.Args {
		le.scope.InsertFuncArg(arg)
		s := e.interner.Intern(arg.Name)
		fargs = append(fargs, data.NewSymbol(s))
	}
	le.emitLiftingForFuncArgs(node.Args, fargs)
	le.emitExprInTailPos(node.Body)
	// todo: implicit return might not always be needed but then
	// we will never get there if there is an explicit one
	le.emitByte(isa.Return)
	e.line = int(node.Beg.Line)
	e.errors = append(e.errors, le.errors...)
	code := le.result
	l := data.NewLambda(name, nil, fargs, code)
	index := e.result.AddConstant(l)
	if index > math.MaxUint16 {
		e.error(node.NodeSpan(), "More constants that uint16 can hold. That is not supported.")
		return
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(isa.Closure)
	e.emitBytes(args...)
}

func (e *Emitter) emitSequence(instr isa.Op, node ast.SequenceLiteral) {
	vals := node.Values()
	for _, expr := range vals {
		e.emitExpr(expr)
	}
	size := len(vals)
	if size > math.MaxUint16 {
		e.error(
			node.NodeSpan(),
			fmt.Sprintf("sequence literals can only support max of %d elements", math.MaxUint16))
		return
	}
	e.line = int(node.NodeSpan().Beg.Line)
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(size))
	e.emitByte(instr)
	e.emitBytes(args...)
}

func (e *Emitter) emitRecord(node *ast.RecordConst) {
	for _, f := range node.Fields {
		e.emitExpr(f.Val)
		e.emitSymbol(f.Key)
	}
	size := len(node.Fields)
	if size > math.MaxUint16 {
		e.error(
			node.NodeSpan(),
			fmt.Sprintf("Record literals can only support max of %d elements", math.MaxUint16))
		return
	}
	e.line = int(node.Beg.Line)
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(size))
	e.emitByte(isa.MakeRecord)
	e.emitBytes(args...)
}

func (e *Emitter) emitLocalEffect(node *ast.LocalEffect) {
	e.emitSymbol(node.Name)
	e.emitByte(isa.MakeEffect)
}

func (e *Emitter) emitAccess(node *ast.Access) {
	e.emitExpr(node.Lhs)
	index, err := e.addSymbol(node.Property.Name)
	if err != nil {
		e.error(node.NodeSpan(), err.Error())
		return
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(isa.GetField)
	e.emitBytes(args...)
}

func (e *Emitter) emitIf(node *ast.IfExpr, tailpos bool) {
	e.emitExpr(node.Cond)
	ifpos := e.emitJumpIfFalse()
	e.emitBlock(node.IfBranch, tailpos)
	if node.ElseBranch != nil {
		skipElse := e.emitJump()
		off := e.result.Len() - ifpos
		e.patchJump(ifpos, off)
		if tailpos {
			e.emitExprInTailPos(node.ElseBranch)
		} else {
			e.emitExpr(node.ElseBranch)
		}
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

func (e *Emitter) emitWhile(node *ast.WhileStmt) {
	lbeg := len(e.result.Instrs)
	e.emitExpr(node.Cond)
	jpos := e.emitJumpIfFalse()
	e.emitStmtBlock(node.Body)
	jb := e.emitJumpBack()
	e.patchJump(jb, jb-lbeg)
	off := e.result.Len() - jpos
	e.patchJump(jpos, off)
}

func (e *Emitter) emitHandler(node *ast.Handle, tailpos bool) {
	if node.HasGuards {
		desnode, effectvars := desugarGuards(node, e.nextCounterVal())
		for _, eff := range effectvars {
			if e.scope.IsGlobal() {
				e.emitGlobalVariableDecl(&ast.GlobalValDecl{
					Span: eff.Span,
					Name: eff.Name,
					Rhs:  eff.Rhs,
				})
			} else {
				e.emitVariableDecl(eff)
			}
		}
		node = desnode
	}
	for _, arm := range node.Arms {
		typ := arm.Effect
		e.emitExpr(typ)
		args := []*ast.FuncDeclArg{{
			Span: arm.Arg.Span,
			Name: arm.Arg.Name,
		}}
		if arm.Continuation != nil {
			args = append(args, &ast.FuncDeclArg{
				Span: arm.Continuation.Span,
				Name: arm.Continuation.Name,
			})
		}
		e.emitLambda(&ast.LambdaExpr{
			Span: arm.Body.Span,
			Args: args,
			Body: arm.Body,
		})
	}
	if len(node.Arms) > math.MaxUint16 {
		e.error(node.Span, "More handler arms than uint16 can hold, that is unsupported")
		return
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(len(node.Arms)))
	e.emitByte(isa.InstallHandler)
	e.emitBytes(args...)
	e.emitLambda(&ast.LambdaExpr{
		Span: node.Body.Span,
		Args: []*ast.FuncDeclArg{},
		Body: node.Body,
	})
	e.emitByte(isa.Call0)
	e.emitByte(isa.PopHandler)
}

func (e *Emitter) emitReturn(node *ast.Return) {
	e.emitExpr(node.Val)
	e.emitByte(isa.Return)
}

func (e *Emitter) emitResume(node *ast.Resume, tailpos bool) {
	e.emitExpr(node.Cont)
	if tailpos {
		// todo: Tail resume only works if it's called in the withclause
		// or if it's called in a function that is tailcalled
		// by the with clause.
		// so clause -> tailcall -> tailcall -> tailresume works
		// but clause ->tailcall -> call -> tailresume does not
		// We can either try to guard against it or maybe instead
		// of "resume" keyword we could create "tailresume" keyword to
		// be more explicit.
		if node.Arg == nil {
			e.emitByte(isa.TailResume0)
		} else {
			e.emitExpr(node.Arg)
			e.emitByte(isa.TailResume1)
		}
	} else {
		e.emitByte(isa.Resume)
		if node.Arg != nil {
			e.emitExpr(node.Arg)
		} else {
			e.emitNone()
		}
		e.emitByte(isa.Call1)
		e.emitByte(isa.PopHandler)
	}
}

// emitStmtBlock emits a list of statements.
// In contrary to normal block this one does
// not push a value at the end.
func (e *Emitter) emitStmtBlock(node *ast.Block) {
	for _, instr := range node.Instr {
		e.emitStmt(instr)
	}
}

func (e *Emitter) emitBlock(node *ast.Block, tailpos bool) {
	for _, instr := range node.Instr[:len(node.Instr)-1] {
		e.emitStmt(instr)
	}
	last := node.Instr[len(node.Instr)-1]
	switch v := last.(type) {
	case *ast.StmtExpr:
		if tailpos {
			e.emitExprInTailPos(v.Expr)
		} else {
			e.emitExpr(v.Expr)
		}
	case *ast.Return:
		e.emitReturn(v)
	default:
		e.emitStmt(v)
		e.emitByte(isa.PushNone)
	}
}

func (e *Emitter) emitExprInTailPos(node ast.Expr) {
	switch v := node.(type) {
	case *ast.Resume:
		e.emitResume(v, true)
	case *ast.IfExpr:
		e.emitIf(v, true)
	case *ast.FuncApplication:
		e.emitApplication(v, true)
	case *ast.Handle:
		e.emitHandler(v, true)
	case *ast.Block:
		e.emitBlock(v, true)
	default:
		e.emitExpr(node)
	}
}

func (e *Emitter) addSymbol(val string) (int, error) {
	s := e.interner.Intern(val)
	v := data.NewSymbol(s)
	index := e.result.AddConstant(v)
	if index > math.MaxUint16 {
		return 0, errors.New("more constants that uint16 can hold, that is not supported")
	}
	return index, nil
}

func (e *Emitter) nextCounterVal() int {
	ret := e.counter
	e.counter++
	return ret
}
