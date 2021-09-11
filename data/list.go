package data

import "strings"

type List struct {
	values []Value
	size   int
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
