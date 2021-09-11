package data

import "strconv"

type Int struct {
	val int
}

func (i Int) String() string {
	return strconv.Itoa(i.val)
}

func (i Int) Equal(o Value) bool {
	if oi, ok := o.(Int); ok {
		return i.val == oi.val
	}
	return false
}
