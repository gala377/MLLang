package data

import (
	"strconv"
)

type Int struct {
	Val int
}

func NewInt(v int) *Int {
	return &Int{v}
}

func (i Int) String() string {
	return strconv.Itoa(i.Val)
}

func (i Int) Equal(o Value) bool {
	if oi, ok := o.(*Int); ok {
		return i.Val == oi.Val
	}
	return false
}
