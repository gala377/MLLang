package data

import "strconv"

type Float struct {
	Val float64
}

func NewFloat(v float64) Float {
	return Float{v}
}

func (f Float) String() string {
	return strconv.FormatFloat(f.Val, 'e', -1, 64)
}

func (f Float) Equal(o Value) bool {
	if of, ok := o.(Float); ok {
		return f.Val == of.Val
	}
	return false
}

func (f Float) Cast(as NumKind) Number {
	switch as {
	case FloatKind:
		return f
	case IntKind:
		return NewInt(int(f.Val))
	}
	panic("unreachable")
}

func (f Float) Kind() NumKind {
	return FloatKind
}

func (f Float) Neg() Value {
	return NewFloat(-f.Val)
}
