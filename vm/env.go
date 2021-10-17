package vm

import "github.com/gala377/MLLang/data"

type Env struct {
	vals map[data.Symbol]data.Value
}

func NewEnv() Env {
	return Env{
		vals: make(map[data.Symbol]data.Value),
	}
}

func EnvFromMap(m map[data.Symbol]data.Value) Env {
	return Env{
		vals: m,
	}
}

func (e *Env) Lookup(s data.Symbol) data.Value {
	return e.vals[s]
}

func (e *Env) Instert(s data.Symbol, v data.Value) {
	e.vals[s] = v
}
