package isa

import "github.com/gala377/MLLang/data"

type Code struct {
	instrs []byte
	consts []data.Value
	lines  []int
}

func NewCode() Code {
	c := Code{
		instrs: make([]byte, 0),
		consts: make([]data.Value, 0),
		lines:  make([]int, 0),
	}
	return c
}

func (c *Code) AddConstant(v data.Value) {
	c.consts = append(c.consts, v)
}
