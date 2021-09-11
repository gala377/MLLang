package data

import (
	"fmt"
)

type Value interface {
	fmt.Stringer
	Equal(Value) bool
}

type noneType struct{}

var None = &noneType{}

func (n *noneType) String() string {
	return "None"
}

func (n *noneType) Equal(o Value) bool {
	_, ok := o.(*noneType)
	return ok
}
