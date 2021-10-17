package data

type InternedString = *string

type Symbol struct {
	name InternedString
}

func NewSymbol(s InternedString) Symbol {
	return Symbol{s}
}

func (s Symbol) String() string {
	return *s.name
}

func (s Symbol) Equal(o Value) bool {
	if os, ok := o.(Symbol); ok {
		return s.name == os.name
	}
	return false
}

func (s Symbol) EqualSymbol(o Symbol) bool {
	return s.name == o.name
}

func (s Symbol) EqualString(o InternedString) bool {
	return s.name == o
}

func (s Symbol) Inner() InternedString {
	return s.name
}
