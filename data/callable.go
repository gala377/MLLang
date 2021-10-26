package data

import (
	"fmt"
	"strings"
)

type (
	ReturnKind = byte

	Trampoline struct {
		Kind ReturnKind
		Code *Code
		Env  *Env
	}
	Callable interface {
		Value
		Arity() int
		Call(...Value) (Value, Trampoline)
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

	Function struct {
		args []Symbol
		name string
		env  *Env
		body *Code
	}
)

const (
	Returned ReturnKind = iota
	Call
	TailCall
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

func (fn *NativeFunc) Call(vv ...Value) (Value, Trampoline) {
	return (fn.fn)(vv...), ProceedTramp
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

func (f *PartialApp) Call(vv ...Value) (Value, Trampoline) {
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

func NewFunction(name string, args []Symbol, body *Code) *Function {
	env := NewEnv()
	return &Function{
		name: name,
		args: args,
		body: body,
		env:  &env,
	}
}

func NewLambda(env *Env, args []Symbol, body *Code) *Function {
	return &Function{
		name: "",
		args: args,
		body: body,
		env:  env,
	}
}

func (f *Function) Arity() int {
	return len(f.args)
}

func (f *Function) Call(vv ...Value) (Value, Trampoline) {
	callenv := make(map[Symbol]Value)
	for k, v := range f.env.Vals {
		callenv[k] = v
	}
	for i, arg := range f.args {
		callenv[arg] = vv[i]
	}
	env := EnvFromMap(callenv)
	t := Trampoline{
		Kind: Call,
		Code: f.body,
		Env:  &env,
	}
	return None, t
}

func (f *Function) String() string {
	name := f.name
	if name == "" {
		name = "lambda"
	}
	return fmt.Sprintf("<function %s>", name)
}

func (f *Function) Equal(o Value) bool {
	if of, ok := o.(*Function); ok {
		return f == of
	}
	return false
}

var ProceedTramp = Trampoline{Kind: Returned}
