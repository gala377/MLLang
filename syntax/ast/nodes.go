package ast

import (
	"fmt"
	"strconv"

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

	GlobalValDecl struct {
		*span.Span
		Name string
		Rhs  Expr
	}

	FuncDecl struct {
		*span.Span
		Name string
		Args []FuncDeclArg
		Body Block
	}

	FuncDeclArg struct {
		*span.Span
		Name string
	}

	Block struct {
		*span.Span
		Instr []Node
	}

	FuncApplication struct {
		*span.Span
		Callee Expr
		Args   []Expr
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
		IfBranch   Block
		ElseBranch Block
	}

	WhileExpr struct {
		*span.Span
		Cond Expr
		Body Block
	}

	LetExpr struct {
		*span.Span
		Decls Expr // Expected to be something that evaluates to a record
		Body  Block
	}
)

func (g *GlobalValDecl) declNode() {}
func (f *FuncDecl) declNode()      {}

func (b *Block) exprNode()           {}
func (f *FuncApplication) exprNode() {}
func (i *IntConst) exprNode()        {}
func (f *FloatConst) exprNode()      {}
func (s *StringConst) exprNode()     {}
func (r *RecordConst) exprNode()     {}
func (l *ListConst) exprNode()       {}
func (t *TupleConst) exprNode()      {}
func (i *IfExpr) exprNode()          {}
func (w *WhileExpr) exprNode()       {}
func (l *LetExpr) exprNode()         {}
func (i *Identifier) exprNode()      {}

func (g *GlobalValDecl) NodeSpan() *span.Span {
	return g.Span
}

func (f *FuncDecl) NodeSpan() *span.Span {
	return f.Span
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

func (w *WhileExpr) NodeSpan() *span.Span {
	return w.Span
}

func (l *LetExpr) NodeSpan() *span.Span {
	return l.Span
}

func (i *Identifier) NodeSpan() *span.Span {
	return i.Span
}

func (g *GlobalValDecl) String() string {
	return fmt.Sprintf(
		`GlobalVar{
	name=%s
	rhs=%s
}`, g.Name, g.Rhs.String())
}

func (f *FuncDecl) String() string {
	return "Unsupported"
}

func (b *Block) String() string {
	return "Unsupported"
}

func (f *FuncApplication) String() string {
	callee := f.Callee.String()
	args := []string{}
	for _, a := range f.Args {
		args = append(args, a.String())
	}
	return fmt.Sprintf("FnApp{callee:{%v}, args:{%v}}", callee, args)
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
	return "Unsupported"
}

func (i *IfExpr) String() string {
	return "Unsupported"
}

func (w *WhileExpr) String() string {
	return "Unsupported"
}

func (l *LetExpr) String() string {
	return "Unsupported"
}

func (i *Identifier) String() string {
	return i.Name
}
