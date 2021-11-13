package data

import (
	"fmt"
	"strings"
)

type Record struct {
	fields map[Symbol]Value
}

func NewRecord(fields map[Symbol]Value) Record {
	return Record{fields: fields}
}

func (r Record) String() string {
	var b strings.Builder
	b.WriteRune('{')
	for key, val := range r.fields {
		b.WriteString(fmt.Sprintf("%v=%v, ", key, val))
	}
	b.WriteRune('}')
	return b.String()
}

func (r Record) Equal(o Value) bool {
	if or, ok := o.(Record); ok {
		return &r.fields == &or.fields
	}
	return false
}
