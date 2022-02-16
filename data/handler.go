package data

import "fmt"

type (
	Type struct {
		Name Symbol
		// Used only when comparing types and for hashing
		Identity *int
	}

	Handler struct {
		// Maps effect types to handlers
		// Handlers have to be functions accepting either one or 2 arguments
		Clauses map[Type]Callable
	}
)

func NewType(name Symbol) Type {
	return Type{
		Name:     name,
		Identity: new(int),
	}
}

func NewHandler(clauses map[Type]Callable) *Handler {
	return &Handler{clauses}
}

func (e Type) String() string {
	return fmt.Sprintf("Effect %s", e.Name)
}

func (e Type) Equal(o Value) bool {
	if ov, ok := o.(Type); ok {
		return e == ov
	}
	return false
}

func (e Type) Arity() int {
	return 1
}

func (e Type) Call(_ VmProxy, vv ...Value) (Value, Trampoline) {
	arg := vv[0]
	return NewTuple([]Value{e, arg}), EffectTramp
}

func (e *Handler) String() string {
	return "Effect Handler"
}

func (e *Handler) Equal(o Value) bool {
	if ov, ok := o.(*Handler); ok {
		return e == ov
	}
	return false
}

func HandlerCapturesContinuation(h Callable) bool {
	return h.Arity() == 2
}
