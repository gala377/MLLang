package codegen

import (
	"encoding/binary"
	"log"
	"math"

	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
	"github.com/gala377/MLLang/syntax/ast"
)

type Emitter struct {
	result   *isa.Code
	line     int
	interner *Interner
}

func NewEmitter() *Emitter {
	c := isa.NewCode()
	e := Emitter{
		result:   &c,
		line:     0,
		interner: NewInterner(),
	}
	return &e
}

func (e *Emitter) Compile(n []ast.Node) *isa.Code {
	e.emitNode(n[0])
	e.emitByte(isa.Return)
	return e.result
}

func (e *Emitter) emitNode(n ast.Node) {
	if v, ok := n.(ast.Expr); ok {
		e.emitExpr(v)
		return
	}
	panic("Not yet supported")
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
	case *ast.Block:
		e.emitBlock(v)
	case *ast.Identifier:
		e.emitLookup(v)
	default:
		log.Printf("Node is %v", node)
		panic("Node cannot be emitted. Not supported")
	}
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
	for _, instr := range node.Instr {
		if v, ok := instr.(ast.Expr); ok {
			e.emitExpr(v)
		} else {
			panic("Emitting node not yet supported")
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

func (e *Emitter) emitLookup(node *ast.Identifier) {
	s := e.interner.Intern(node.Name)
	v := data.NewSymbol(s)
	index := e.result.AddConstant(v)
	if index > math.MaxUint16 {
		panic("More constants that uint16 can hold. That is not supported.")
	}
	args := []byte{0, 0}
	binary.BigEndian.PutUint16(args, uint16(index))
	e.emitByte(isa.DynLookup)
	e.emitBytes(args...)
}
