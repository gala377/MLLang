package isa

import "github.com/gala377/MLLang/data"

type Code struct {
	Instrs []byte
	Consts []data.Value
	// todo: change Lines to something like runing sum encoding or so.
	Lines []int
}

func NewCode() Code {
	c := Code{
		Instrs: make([]byte, 0),
		Consts: make([]data.Value, 0),
		Lines:  make([]int, 0),
	}
	return c
}

func (c *Code) AddConstant(v data.Value) int {
	c.Consts = append(c.Consts, v)
	return len(c.Consts) - 1
}

func (c *Code) WriteByte(b byte, line int) {
	c.Instrs = append(c.Instrs, b)
	c.Lines = append(c.Lines, line)
}

func (c *Code) ReadByte(offset int) byte {
	return c.Instrs[offset]
}

func (c *Code) GetConstant(i byte) data.Value {
	return c.Consts[i]
}

func (c *Code) Len() int {
	return len(c.Instrs)
}
