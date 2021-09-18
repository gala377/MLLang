package isa

import "github.com/gala377/MLLang/data"

type Code struct {
	instrs []byte
	consts []data.Value
}
