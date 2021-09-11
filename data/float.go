package data

import "strconv"

type Float struct {
	val float64
}

func (f Float) String() string {
	return strconv.FormatFloat(f.val, 'e', -1, 64)
}

func (f Float) Equal(o Value) bool {
	if of, ok := o.(Float); ok {
		return f.val == of.val
	}
	return false
}
