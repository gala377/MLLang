package data

import (
	"fmt"
	"strings"
)

type (
	Callable interface {
		Value
		Arity() int
		Call(...Value) Value
	}

	NativeFunc struct {
		fn    func(...Value) Value
		arity int
		name  string
	}

	PartialApp struct {
		args []Value
		fn   Callable
	}
)

func NewNativeFunc(name string, arity int, fn func(...Value) Value) NativeFunc {
	return NativeFunc{
		name:  name,
		arity: arity,
		fn:    fn,
	}
}

func (fn *NativeFunc) Arity() int {
	return fn.arity
}

func (fn *NativeFunc) Call(vv ...Value) Value {
	return (fn.fn)(vv...)
}

func (fn *NativeFunc) String() string {
	return fmt.Sprintf("<Native func %s>", fn.name)
}

func (fn *NativeFunc) Equal(o Value) bool {
	if of, ok := o.(*NativeFunc); ok {
		return fn == of
	}
	return false
}

func NewPartialApp(c Callable, vv ...Value) *PartialApp {
	if len(vv) >= c.Arity() {
		panic("Can only apply partially if there are less arguments than the function needs")
	}
	if p, ok := c.(*PartialApp); ok {
		p.args = append(p.args, vv...)
		return p
	}
	return &PartialApp{
		args: vv,
		fn:   c,
	}
}

func (f *PartialApp) Arity() int {
	return f.fn.Arity() - len(f.args)
}

func (f *PartialApp) Call(vv ...Value) Value {
	args := append(f.args, vv...)
	return f.fn.Call(args...)
}

func (f *PartialApp) String() string {
	vals := make([]string, 0, len(f.args))
	for _, arg := range f.args {
		vals = append(vals, arg.String())
	}
	return fmt.Sprintf("<Partial app %s %s>", f.fn.String(), strings.Join(vals, " "))
}

func (f *PartialApp) Equal(o Value) bool {
	if of, ok := o.(*PartialApp); ok {
		return f == of
	}
	return false
}
