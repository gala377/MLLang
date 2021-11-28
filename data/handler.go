package data

import "fmt"

type (
	EffectType struct {
		Name Symbol
	}

	Handler struct {
		// Maps effect types to handlers
		// Handlers have to functions accepting either one or 2 arguments
		Clauses map[EffectType]Value
	}
)

func NewEffectType(name Symbol) *EffectType {
	return &EffectType{name}
}

func NewHandler(clauses map[EffectType]Value) *Handler {
	return &Handler{clauses}
}

func (e *EffectType) String() string {
	return fmt.Sprintf("Effect %s", e.Name)
}

func (e *EffectType) Equal(o Value) bool {
	if ov, ok := o.(*EffectType); ok {
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
