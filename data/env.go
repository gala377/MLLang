package data

type Env struct {
	Vals map[Symbol]Value
}

func NewEnv() Env {
	return Env{
		Vals: make(map[Symbol]Value),
	}
}

func EnvFromMap(m map[Symbol]Value) Env {
	return Env{
		Vals: m,
	}
}

func (e *Env) Lookup(s Symbol) Value {
	return e.Vals[s]
}

func (e *Env) Insert(s Symbol, v Value) {
	e.Vals[s] = v
}

func (e *Env) String() string {
	return "runtime environment"
}

func (e *Env) Equal(o Value) bool {
	panic("environment values should never be compared")
}
