package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gala377/MLLang/syntax/span"
)

type (
	Node interface {
		fmt.Stringer
		NodeSpan() *span.Span
	}

	Decl interface {
		Node
		declNode()
	}

	Expr interface {
		Node
		exprNode()
	}

	Stmt interface {
		Node
		stmtNode()
	}

	GlobalValDecl struct {
		*span.Span
		Name string
		Rhs  Expr
	}

	ValDecl struct {
		*span.Span
		Name string
		Rhs  Expr
	}

	StmtExpr struct {
		Expr
	}

	FuncDecl struct {
		*span.Span
		Name string
		Args []FuncDeclArg
		Body Expr
	}

	FuncDeclArg struct {
		*span.Span
		Name string
	}

	Block struct {
		*span.Span
		Instr []Stmt
	}

	FuncApplication struct {
		*span.Span
		Callee Expr
		Args   []Expr
		Block  *LambdaExpr
	}

	IntConst struct {
		*span.Span
		Val int
	}

	FloatConst struct {
		*span.Span
		Val float64
	}

	StringConst struct {
		*span.Span
		Val string
	}

	Identifier struct {
		*span.Span
		Name string
	}
	RecordConst struct {
		*span.Span
		Fields map[string]Expr
	}
	ListConst struct {
		*span.Span
		Vals []Expr
	}
	TupleConst struct {
		*span.Span
		Vals []Expr
	}

	IfExpr struct {
		*span.Span
		Cond       Expr
		IfBranch   *Block
		ElseBranch Expr
	}

	WhileStmt struct {
		*span.Span
		Cond Expr
		Body *Block
	}

	LetExpr struct {
		*span.Span
		Decls Expr // Expected to be something that evaluates to a record
		Body  *Block
	}

	LambdaExpr struct {
		*span.Span
		Args []FuncDeclArg
		Body Expr
	}

	BoolConst struct {
		*span.Span
		Val bool
	}

	Assignment struct {
		*span.Span
		LValue Expr
		RValue Expr
	}
)

func (g *GlobalValDecl) declNode() {}
func (f *FuncDecl) declNode()      {}

func (v *ValDecl) stmtNode()    {}
func (w *WhileStmt) stmtNode()  {}
func (s *StmtExpr) stmtNode()   {}
func (a *Assignment) stmtNode() {}

func (b *Block) exprNode()           {}
func (f *FuncApplication) exprNode() {}
func (i *IntConst) exprNode()        {}
func (f *FloatConst) exprNode()      {}
func (s *StringConst) exprNode()     {}
func (r *RecordConst) exprNode()     {}
func (l *ListConst) exprNode()       {}
func (t *TupleConst) exprNode()      {}
func (i *IfExpr) exprNode()          {}
func (l *LetExpr) exprNode()         {}
func (i *Identifier) exprNode()      {}
func (l *LambdaExpr) exprNode()      {}
func (b *BoolConst) exprNode()       {}

func (g *GlobalValDecl) NodeSpan() *span.Span {
	return g.Span
}

func (f *FuncDecl) NodeSpan() *span.Span {
	return f.Span
}

func (v *ValDecl) NodeSpan() *span.Span {
	return v.Span
}

func (s *StmtExpr) NodeSpan() *span.Span {
	return s.Expr.NodeSpan()
}

func (b *Block) NodeSpan() *span.Span {
	return b.Span
}

func (f *FuncApplication) NodeSpan() *span.Span {
	return f.Span
}

func (i *IntConst) NodeSpan() *span.Span {
	return i.Span
}

func (f *FloatConst) NodeSpan() *span.Span {
	return f.Span
}

func (s *StringConst) NodeSpan() *span.Span {
	return s.Span
}

func (r *RecordConst) NodeSpan() *span.Span {
	return r.Span
}

func (l *ListConst) NodeSpan() *span.Span {
	return l.Span
}

func (t *TupleConst) NodeSpan() *span.Span {
	return t.Span
}

func (i *IfExpr) NodeSpan() *span.Span {
	return i.Span
}

func (w *WhileStmt) NodeSpan() *span.Span {
	return w.Span
}

func (l *LetExpr) NodeSpan() *span.Span {
	return l.Span
}

func (i *Identifier) NodeSpan() *span.Span {
	return i.Span
}

func (l *LambdaExpr) NodeSpan() *span.Span {
	return l.Span
}

func (b *BoolConst) NodeSpan() *span.Span {
	return b.Span
}

func (a *Assignment) NodeSpan() *span.Span {
	return a.Span
}

func (g *GlobalValDecl) String() string {
	return fmt.Sprintf(
		`GlobalVar{
	name=%s
	rhs=%s
}`, g.Name, g.Rhs)
}

func (f *FuncDecl) String() string {
	msg := "FnDecl{" + f.Name
	for _, arg := range f.Args {
		msg += " " + arg.Name
	}
	msg += "} "
	msg += f.Body.String()
	return msg
}

func (b *Block) String() string {
	msg := "Block:\n"
	for _, instr := range b.Instr {
		msg += fmt.Sprintf("\t%v\n", instr)
	}
	return msg
}

func (g *ValDecl) String() string {
	return fmt.Sprintf(
		`Var{
	name=%s
	rhs=%s
}`, g.Name, g.Rhs)
}

func (s *StmtExpr) String() string {
	return fmt.Sprintf("Stmt{%s}", s.Expr)
}

func (f *FuncApplication) String() string {
	callee := f.Callee.String()
	args := []string{}
	for _, a := range f.Args {
		args = append(args, a.String())
	}
	block := ""
	if f.Block != nil {
		block = fmt.Sprintf(" Block %s", f.Block)
	}
	return fmt.Sprintf("FnApp{callee:{%v}, args:{%v}}%s", callee, args, block)
}

func (i *IntConst) String() string {
	return fmt.Sprintf("IntConst{%v}", i.Val)
}

func (f *FloatConst) String() string {
	return strconv.FormatFloat(f.Val, 'e', -1, 64)
}

func (s *StringConst) String() string {
	return "Unsupported"
}

func (r *RecordConst) String() string {
	return "Unsupported"
}

func (l *ListConst) String() string {
	return "Unsupported"
}

func (t *TupleConst) String() string {
	vals := make([]string, 0, len(t.Vals))
	for _, val := range t.Vals {
		vals = append(vals, val.String())
	}
	sv := strings.Join(vals, ", ")
	return fmt.Sprintf("Tuple{%s}", sv)
}

func (i *IfExpr) String() string {
	msg := fmt.Sprintf("If{%s} %s", i.Cond, i.IfBranch)
	if i.ElseBranch != nil {
		msg += fmt.Sprintf("\nElse %s", i.ElseBranch)
	}
	return msg
}

func (w *WhileStmt) String() string {
	return fmt.Sprintf("While{%s} %s", w.Cond, w.Body)
}

func (l *LetExpr) String() string {
	return "Unsupported"
}

func (i *Identifier) String() string {
	return i.Name
}

func (l *LambdaExpr) String() string {
	msg := "Lambda{"
	for _, arg := range l.Args {
		msg += " " + arg.Name
	}
	msg += "} "
	msg += l.Body.String()
	return msg
}

func (b *BoolConst) String() string {
	return fmt.Sprintf("%v", b.Val)
}

func (a *Assignment) String() string {
	return fmt.Sprintf(
		"Assignment{\n\tLval=%s,\nRval=%s\n}",
		a.LValue, a.RValue)
}
