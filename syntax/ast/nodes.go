package ast

import (
	"github.com/gala377/MLLang/syntax/span"
)

type (
	Node interface {
		BegPos() span.Position
		EndPos() span.Position
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
		Name string
		Rhs  Expr
	}

	FuncDecl struct {
		Name string
		Args []FuncDeclArg
		Body Block
	}

	FuncDeclArg struct {
		Name string
	}

	Block struct {
		Instr []Node
	}

	FuncApplication struct {
		Callee string
		Args   []Expr
	}

	IntConst struct {
		Val int
	}

	FloatConst struct {
		Val float64
	}

	StringConst struct {
		Val string
	}
)
