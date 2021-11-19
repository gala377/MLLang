package data

import (
	"fmt"
	"strings"
)

type Record struct {
	fields map[Symbol]Value
	keys   []Symbol
}

func RecordFromMap(fields map[Symbol]Value) *Record {
	keys := make([]Symbol, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	return &Record{fields, keys}
}

func EmptyRecord() *Record {
	return &Record{
		fields: make(map[Symbol]Value),
		keys:   make([]Symbol, 0),
	}
}

func (r *Record) String() string {
	var b strings.Builder
	b.WriteRune('{')
	for _, key := range r.keys {
		b.WriteString(fmt.Sprintf("%v=%v, ", key, r.fields[key]))
	}
	b.WriteRune('}')
	return b.String()
}

func (r *Record) Equal(o Value) bool {
	if or, ok := o.(*Record); ok {
		return &r.fields == &or.fields
	}
	return false
}

func (r *Record) GetField(s Symbol) (Value, bool) {
	v, ok := r.fields[s]
	return v, ok
}

func (r *Record) SetField(s Symbol, v Value) {
	_, ok := r.fields[s]
	if !ok {
		r.keys = append(r.keys, s)
	}
	r.fields[s] = v
}

func (r *Record) Get(i Int) (Value, error) {
	idx := i.Val
	if len(r.keys) <= idx {
		return nil, fmt.Errorf("record index out of range idx=%d, size=%d", idx, len(r.keys))
	}
	k := r.keys[idx]
	v := r.fields[k]
	t := NewTuple([]Value{k, v})
	return t, nil
}

func (r *Record) Append(v Value) error {
	at, ok := v.(*Tuple)
	err := fmt.Errorf("records has to append tuples (symbol, value)")
	if !ok {
		return err
	}
	if at.size != 2 {
		return err
	}
	k, ok := at.values[0].(Symbol)
	if !ok {
		return err
	}
	_, ok = r.fields[k]
	if ok {
		return fmt.Errorf("record already has field %s", k)
	}
	nv := at.values[1]
	r.keys = append(r.keys, k)
	r.fields[k] = nv
	return nil
}

func (r *Record) Len() int {
	return len(r.keys)
}
