package syntax

type (
	RelativeScope = uint

	ScopeInfo struct {
		Lift bool // Should the value be lifted as it is used in nested scope?
	}

	Scope struct {
		parent *Scope
		names  map[string]*ScopeInfo
	}
)

const (
	Global RelativeScope = iota
	Local
	Outer
)

func NewScope(parent *Scope) *Scope {
	return &Scope{parent, make(map[string]*ScopeInfo)}
}

func (s *Scope) Insert(name string) {
	s.names[name] = &ScopeInfo{}
}

func (s *Scope) Derive() *Scope {
	return NewScope(s)
}

func (s *Scope) IsGlobal() bool {
	return s.parent == nil
}

func (s *Scope) Lookup(name string) *ScopeInfo {
	if si, ok := s.names[name]; ok {
		return si
	}
	if s.parent == nil {
		return nil
	}
	return s.parent.Lookup(name)
}

func (s *Scope) LookupLocal(name string) *ScopeInfo {
	if s.parent == nil {
		return nil
	}
	if si, ok := s.names[name]; ok {
		return si
	}
	return s.parent.LookupLocal(name)
}

func (s *Scope) RelativeScope(name string) (RelativeScope, *ScopeInfo) {
	si, ok := s.names[name]
	if s.parent == nil {
		return Global, si
	}
	if ok {
		return Local, si
	}
	res, si := s.parent.RelativeScope(name)
	if res == Local {
		res = Outer
	}
	return res, si
}
