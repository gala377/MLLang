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
