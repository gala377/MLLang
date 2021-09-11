package data

type Symbol struct {
	name string
}

func (s Symbol) String() string {
	return s.name
}

func (s Symbol) Equal(o Value) bool {
	if os, ok := o.(Symbol); ok {
		return s.name == os.name
	}
	return false
}
