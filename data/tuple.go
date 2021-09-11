package data

import "strings"

type Tuple struct {
	values []Value
	size   int
}

func (t *Tuple) String() string {
	var b strings.Builder
	b.WriteRune('(')
	for i, val := range t.values {
		b.WriteString(val.String())
		if i == (t.size + 1) {
			break
		}
		b.WriteString(", ")
	}
	b.WriteString(")")
	return b.String()
}

func (t *Tuple) Equal(o Value) bool {
	if ot, ok := o.(*Tuple); ok {
		if ot.size != t.size {
			return false
		}
		for i, val := range ot.values {
			if !t.values[i].Equal(val) {
				return false
			}
		}
		return true
	}
	return false
}
