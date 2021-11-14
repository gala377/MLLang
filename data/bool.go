package data

import "fmt"

type Bool struct {
	Val bool
}

func NewBool(v bool) Bool {
	return Bool{v}
}

func (b Bool) String() string {
	return fmt.Sprintf("%v", b.Val)
}

func (b Bool) Equal(o Value) bool {
	if ob, ok := o.(Bool); ok {
		return b.Val == ob.Val
	}
	return false
}
