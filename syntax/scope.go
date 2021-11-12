package syntax

import "github.com/gala377/MLLang/syntax/ast"

type (
	RelativeScope = uint

	ScopeInfo interface {
		Lift()
		IsLifted() bool
	}

	Scope struct {
		parent *Scope
		names  map[string]ScopeInfo
	}

	emptyScopeInfo struct{}
	varScopeInfo   struct {
		inner *ast.ValDecl
	}
	fnArgScopeInfo struct {
		inner *ast.FuncDeclArg
	}
)

const (
	Global RelativeScope = iota
	Local
	Outer
)

func (e emptyScopeInfo) Lift() {}
func (e emptyScopeInfo) IsLifted() bool {
	return false
}
func (v varScopeInfo) Lift() {
	v.inner.Lift = true
}
func (v varScopeInfo) IsLifted() bool {
	return v.inner.Lift
}
func (a fnArgScopeInfo) Lift() {
	a.inner.Lift = true
}
func (a fnArgScopeInfo) IsLifted() bool {
	return a.inner.Lift
}

func NewScope(parent *Scope) *Scope {
	return &Scope{parent, make(map[string]ScopeInfo)}
}

func (s *Scope) Insert(name string) {
	s.names[name] = emptyScopeInfo{}
}

func (s *Scope) InsertVal(decl *ast.ValDecl) {
	s.names[decl.Name] = varScopeInfo{decl}
}

func (s *Scope) InsertFuncArg(arg *ast.FuncDeclArg) {
	s.names[arg.Name] = fnArgScopeInfo{arg}
}

func (s *Scope) Derive() *Scope {
	return NewScope(s)
}

func (s *Scope) IsGlobal() bool {
	return s.parent == nil
}

func (s *Scope) Lookup(name string) ScopeInfo {
	if si, ok := s.names[name]; ok {
		return si
	}
	if s.parent == nil {
		return nil
	}
	return s.parent.Lookup(name)
}

func (s *Scope) LookupLocal(name string) ScopeInfo {
	if s.parent == nil {
		return nil
	}
	if si, ok := s.names[name]; ok {
		return si
	}
	return s.parent.LookupLocal(name)
}

func (s *Scope) RelativeScope(name string) (RelativeScope, ScopeInfo) {
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
