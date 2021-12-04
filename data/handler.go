package data

import "fmt"

type (
	Type struct {
		Name Symbol
	}

	Handler struct {
		// Maps effect types to handlers
		// Handlers have to be functions accepting either one or 2 arguments
		Clauses map[*Type]Value
	}
)

func NewType(name Symbol) *Type {
	return &Type{name}
}

func NewHandler(clauses map[*Type]Value) *Handler {
	return &Handler{clauses}
}

func (e *Type) String() string {
	return fmt.Sprintf("Effect %s", e.Name)
}

func (e *Type) Equal(o Value) bool {
	if ov, ok := o.(*Type); ok {
		return e == ov
	}
	return false
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
