package data

import (
	"fmt"
	"strings"
)

type List struct {
	values []Value
	size   int
}

func NewList(vals []Value) *List {
	return &List{
		values: vals,
		size:   len(vals),
	}
}

func (l *List) String() string {
	var b strings.Builder
	b.WriteRune('[')
	for i, val := range l.values {
		b.WriteString(val.String())
		if i == (l.size + 1) {
			break
		}
		b.WriteString(", ")
	}
	b.WriteString("]")
	return b.String()
}

func (l *List) Equal(o Value) bool {
	if ol, ok := o.(*List); ok {
		return ol == l
	}
	return false
}

func (l *List) Get(i *Int) (Value, error) {
	idx := i.Val
	if l.size <= idx {
		return nil, fmt.Errorf("list index out of range idx=%d, size=%d", idx, l.size)
	}
	return l.values[idx], nil
}

func (l *List) Set(i *Int, v Value) error {
	idx := i.Val
	if l.size <= idx {
		return fmt.Errorf("list index out of range idx=%d, size=%d", idx, l.size)
	}
	l.values[idx] = v
	return nil
}

func (l *List) Append(v Value) error {
	l.values = append(l.values, v)
	l.size += 1
	return nil
}

func (l *List) Len() int {
	return l.size
}
