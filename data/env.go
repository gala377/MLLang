package data

import (
	"fmt"
	"sync"
)

type Env struct {
	lock sync.RWMutex
	Vals map[Symbol]Value
}

func NewEnv() *Env {
	return &Env{
		Vals: make(map[Symbol]Value),
	}
}

func EnvFromMap(m map[Symbol]Value) *Env {
	return &Env{
		Vals: m,
	}
}

func (e *Env) Lookup(s Symbol) Value {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.Vals[s]
}

func (e *Env) Insert(s Symbol, v Value) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.Vals[s] = v
}

func (e *Env) Set(s Symbol, v Value) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	_, ok := e.Vals[s]
	if !ok {
		return fmt.Errorf("unknown name %s", s)
	}
	e.Vals[s] = v
	return nil
}

func (e *Env) String() string {
	return "runtime environment"
}

func (e *Env) Equal(o Value) bool {
	panic("environment values should never be compared")
}

func (e *Env) Clone() *Env {
	e.lock.RLock()
	defer e.lock.RUnlock()
	vv := make(map[Symbol]Value)
	for k, v := range e.Vals {
		vv[k] = v
	}
	return &Env{Vals: vv}
}
