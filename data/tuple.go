package data

import (
	"fmt"
	"strings"
)

type Tuple struct {
	values []Value
	size   int
}

func NewTuple(vals []Value) *Tuple {
	return &Tuple{
		values: vals,
		size:   len(vals),
	}
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

func (t *Tuple) Get(i *Int) (Value, error) {
	idx := i.Val
	if t.size <= idx {
		return nil, fmt.Errorf("tuple index out of range idx=%d, size=%d", idx, t.size)
	}
	return t.values[idx], nil
}

func (t *Tuple) Len() int {
	return t.size
}
