package codegen

import "github.com/gala377/MLLang/data"

type Interner struct {
	symbols []data.InternedString
	mapper  map[string]int
}

func NewInterner() *Interner {
	return &Interner{
		symbols: make([]data.InternedString, 0),
		mapper:  make(map[string]int),
	}
}

func (ir *Interner) Intern(s string) data.InternedString {
	if i, ok := ir.mapper[s]; ok {
		return ir.symbols[i]
	}
	i := len(ir.symbols)
	ir.symbols = append(ir.symbols, &s)
	ir.mapper[s] = i
	return &s
}
