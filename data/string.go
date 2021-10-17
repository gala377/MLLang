package data

import "fmt"

type String struct {
	Val string
}

func NewString(v string) String {
	return String{v}
}

func (s String) String() string {
	return fmt.Sprintf("\"%s\"", s.Val)
}

func (s String) Equal(o Value) bool {
	if os, ok := o.(String); ok {
		return s.Val == os.Val
	}
	return false
}
