package data

import (
	"strconv"
)

type Int struct {
	Val int
}

func NewInt(v int) Int {
	return Int{v}
}

func (i Int) String() string {
	return strconv.Itoa(i.Val)
}

func (i Int) Equal(o Value) bool {
	if oi, ok := o.(Int); ok {
		return i.Val == oi.Val
	}
	return false
}

func (i Int) Cast(as NumKind) Number {
	switch as {
	case IntKind:
		return i
	case FloatKind:
		return NewFloat(float64(i.Val))
	}
	panic("unreachable")
}

func (i Int) Kind() NumKind {
	return IntKind
}

func (i Int) Neg() Value {
	return NewInt(-i.Val)
}
