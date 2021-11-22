package data

import (
	"fmt"
	"strings"
)

type Tuple struct {
	values []Value
}

func NewTuple(vals []Value) Tuple {
	return Tuple{
		values: vals,
	}
}

func (t Tuple) String() string {
	var b strings.Builder
	b.WriteRune('(')
	for i, val := range t.values {
		b.WriteString(val.String())
		if i == (len(t.values) - 1) {
			break
		}
		b.WriteString(", ")
	}
	b.WriteString(")")
	return b.String()
}

func (t Tuple) Equal(o Value) bool {
	if ot, ok := o.(Tuple); ok {
		if t.Len() != ot.Len() {
			return false
		}
		// Equal if elements are equal
		for i, val := range ot.values {
			if !t.values[i].Equal(val) {
				return false
			}
		}
		return true
	}
	return false
}

func (t Tuple) Get(i Int) (Value, error) {
	idx := i.Val
	if len(t.values) <= idx {
		return nil, fmt.Errorf("tuple index out of range idx=%d, size=%d", idx, t.Len())
	}
	return t.values[idx], nil
}

func (t Tuple) Len() int {
	return len(t.values)
}
