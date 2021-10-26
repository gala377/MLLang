package codegen

type Scope struct {
	parent *Scope
	names  map[string]struct{}
}

func NewScope(parent *Scope) *Scope {
	return &Scope{parent, make(map[string]struct{})}
}

func (s *Scope) Insert(name string) {
	s.names[name] = struct{}{}
}

func (s *Scope) Derive() *Scope {
	return NewScope(s)
}

func (s *Scope) IsGlobal() bool {
	return s.parent == nil
}

func (s *Scope) Lookup(name string) bool {
	if _, ok := s.names[name]; ok {
		return true
	}
	if s.parent == nil {
		return false
	}
	return s.parent.Lookup(name)
}

func (s *Scope) LookupLocal(name string) bool {
	if _, ok := s.names[name]; ok {
		return s.parent != nil
	}
	if s.parent == nil {
		return false
	}
	return s.parent.LookupLocal(name)
}
