package data

import (
	"fmt"
	"strings"
)

type (
	ReturnKind = byte

	Trampoline struct {
		Kind ReturnKind
		Ip   int
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
		SourceLine() int
		FileName() string
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

	Continuation struct {
		Handler *Handler
		Code    *Code
		Ip      int
		Env     *Env
		Stack   []Value
	}
)

const (
	// Immedietaly returns with value returned with trampoline
	Returned ReturnKind = iota
	// Executes code with environment returned in the trampoline
	Call
	TailCall
	// Performs an effect, value returned alongside this trampline
	// has to be a tuple of (effect type, effect value)
	Effect
	// Restores continuation by extending the stack with the list
	// of values returned alongside this continuation.
	RestoreContinuation
	// Immediately executes error in the vm with the message
	// taken from the value returned alongside this trampoline
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

func NewPartialApp(c Callable, vv ...Value) PartialApp {
	if len(vv) >= c.Arity() {
		panic("Can only apply partially if there are less arguments than the function needs")
	}
	if p, ok := c.(PartialApp); ok {
		args := make([]Value, len(p.args), len(p.args)+len(vv))
		copy(args, p.args)
		args = append(args, vv...)
		return PartialApp{
			args: args,
			fn:   p.fn,
		}
	}
	return PartialApp{
		args: vv,
		fn:   c,
	}
}

func PartialApp1(c Callable, arg Value) PartialApp {
	if c.Arity() <= 1 {
		panic("ICE: parial application on function that should be called")
	}
	if p, ok := c.(PartialApp); ok {
		args := make([]Value, len(p.args), len(p.args)+1)
		copy(args, p.args)
		args = append(args, arg)
		return PartialApp{
			args: args,
			fn:   p.fn,
		}
	}
	return PartialApp{
		args: []Value{arg},
		fn:   c,
	}
}

func (f PartialApp) Arity() int {
	return f.fn.Arity() - len(f.args)
}

func (f PartialApp) Call(vm VmProxy, vv ...Value) (Value, Trampoline) {
	args := append(f.args, vv...)
	return f.fn.Call(vm, args...)
}

func (f PartialApp) String() string {
	vals := make([]string, 0, len(f.args))
	for _, arg := range f.args {
		vals = append(vals, arg.String())
	}
	return fmt.Sprintf("<Partial app %s %s>", f.fn.String(), strings.Join(vals, " "))
}

func (f PartialApp) Equal(o Value) bool {
	if of, ok := o.(PartialApp); ok {
		if len(f.args) != len(of.args) {
			return false
		}
		if len(f.args) == 0 {
			return f.fn.Equal(of)
		}
		return &f.args[1] == &of.args[1] && f.fn.Equal(of)
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

func NewContinuation(stack []Value, handler *Handler, ip int, code *Code, env *Env) *Continuation {
	return &Continuation{
		Stack:   stack,
		Handler: handler,
		Code:    code,
		Ip:      ip,
		Env:     env,
	}
}

func (c *Continuation) Arity() int {
	return 1
}

func (c *Continuation) Call(_ VmProxy, vv ...Value) (Value, Trampoline) {
	// return arg and stored stack
	ret := NewTuple([]Value{vv[0], NewList(c.Stack)})
	return ret, Trampoline{
		Kind: RestoreContinuation,
		Ip:   c.Ip,
		Env:  c.Env,
		Code: c.Code,
	}
}

func (c *Continuation) Equal(o Value) bool {
	if ov, ok := o.(*Continuation); ok {
		return ov == c
	}
	return false
}

func (c *Continuation) String() string {
	return "<captured continuation>"
}

var ReturnTramp = Trampoline{Kind: Returned}
var ErrorTramp = Trampoline{Kind: Error}
var EffectTramp = Trampoline{Kind: Effect}
