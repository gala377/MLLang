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
		Call(VmProxy, ...Value) (Value, Trampoline)
	}

	VmProxy interface {
		CreateSymbol(string) Symbol
		GenerateSymbol() Symbol
		Panic(string)
		Clone() VmProxy
		RunClosure(Callable, ...Value) Value
	}
	NativeFunc struct {
		fn    func(VmProxy, ...Value) (Value, error)
		arity int
		name  string
	}

	PartialApp struct {
		args []Value
		fn   Callable
	}

	Closure struct {
		Args []Symbol
		Name Symbol
		Env  *Env
		Body *Code
	}
)

const (
	Returned ReturnKind = iota
	Call
	TailCall
	Error
)

func NewNativeFunc(name string, arity int, fn func(VmProxy, ...Value) (Value, error)) *NativeFunc {
	return &NativeFunc{
		name:  name,
		arity: arity,
		fn:    fn,
	}
}

func (fn *NativeFunc) Arity() int {
	return fn.arity
}

func (fn *NativeFunc) Call(vm VmProxy, vv ...Value) (Value, Trampoline) {
	res, err := (fn.fn)(vm, vv...)
	if err != nil {
		return NewString(err.Error()), ErrorTramp
	}
	return res, ReturnTramp
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

func (f *PartialApp) Call(vm VmProxy, vv ...Value) (Value, Trampoline) {
	args := append(f.args, vv...)
	return f.fn.Call(vm, args...)
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

func NewFunction(name Symbol, args []Symbol, body *Code) *Closure {
	env := NewEnv()
	return &Closure{
		Name: name,
		Args: args,
		Body: body,
		Env:  env,
	}
}

func NewLambda(name Symbol, env *Env, args []Symbol, body *Code) *Closure {
	return &Closure{
		Name: name,
		Args: args,
		Body: body,
		Env:  env,
	}
}

func (f *Closure) Arity() int {
	return len(f.Args)
}

func (f *Closure) Call(_ VmProxy, vv ...Value) (Value, Trampoline) {
	callenv := make(map[Symbol]Value)
	for k, v := range f.Env.Vals {
		callenv[k] = v
	}
	for i, arg := range f.Args {
		callenv[arg] = vv[i]
	}
	if f.Name.Inner() != nil {
		callenv[f.Name] = f
	}
	env := EnvFromMap(callenv)
	t := Trampoline{
		Kind: Call,
		Code: f.Body,
		Env:  env,
	}
	return None, t
}

func (f *Closure) String() string {
	if f.Name.Inner() == nil {
		return "<anonymous function>"
	}
	return fmt.Sprintf("<function %s>", *f.Name.Inner())
}

func (f *Closure) Equal(o Value) bool {
	if of, ok := o.(*Closure); ok {
		return f == of
	}
	return false
}

var ReturnTramp = Trampoline{Kind: Returned}
var ErrorTramp = Trampoline{Kind: Error}
