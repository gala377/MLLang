package ast

type NodeEqual interface {
	Node
	Equal(Node) bool
}

func AstEqual(n1 Node, n2 Node) bool {
	if n1 == nil {
		return n2 == nil
	}
	cmp, ok := n1.(NodeEqual)
	if !ok {
		return false
	}
	return cmp.Equal(n2)
}

func (g *GlobalValDecl) Equal(o Node) bool {
	if og, ok := o.(*GlobalValDecl); ok {
		if og.Name != g.Name {
			return false
		}
		return AstEqual(g.Rhs, og.Rhs)
	}
	return false
}

func (f *FuncDecl) Equal(o Node) bool {
	if of, ok := o.(*FuncDecl); ok {
		if f.Name != of.Name {
			return false
		}
		if len(f.Args) != len(of.Args) {
			return false
		}
		for i, arg := range f.Args {
			if arg.Name != of.Args[i].Name {
				return false
			}
		}
		return AstEqual(f.Body, of.Body)
	}
	return false
}

func (v *ValDecl) Equal(o Node) bool {
	if ov, ok := o.(*ValDecl); ok {
		if ov.Name != v.Name {
			return false
		}
		return AstEqual(v.Rhs, ov.Rhs)
	}
	return false
}

func (s *StmtExpr) Equal(o Node) bool {
	if os, ok := o.(*StmtExpr); ok {
		return AstEqual(s.Expr, os.Expr)
	}
	return false
}

func (b *Block) Equal(o Node) bool {
	if ob, ok := o.(*Block); ok {
		if len(b.Instr) != len(ob.Instr) {
			return false
		}
		for i, inst := range b.Instr {
			if !AstEqual(inst, ob.Instr[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (f *FuncApplication) Equal(o Node) bool {
	if of, ok := o.(*FuncApplication); ok {
		if !AstEqual(f.Callee, of.Callee) {
			return false
		}
		if len(f.Args) != len(of.Args) {
			return false
		}
		for i, a := range f.Args {
			if !AstEqual(a, of.Args[i]) {
				return false
			}
		}
		if f.Block != nil {
			if of.Block == nil {
				return false
			}
			return f.Block.Equal(of.Block)
		}
		return of.Block == nil
	}
	return false
}

func (i *IntConst) Equal(o Node) bool {
	if oi, ok := o.(*IntConst); ok {
		return i.Val == oi.Val
	}
	return false
}

func (f *FloatConst) Equal(o Node) bool {
	if of, ok := o.(*FloatConst); ok {
		return f.Val == of.Val
	}
	return false
}

func (s *StringConst) Equal(o Node) bool {
	if os, ok := o.(*StringConst); ok {
		return s.Val == os.Val
	}
	return false
}

func (l *ListConst) Equal(o Node) bool {
	if ol, ok := o.(*ListConst); ok {
		if len(l.Vals) != len(ol.Vals) {
			return false
		}
		for i, v := range l.Vals {
			if !AstEqual(v, l.Vals[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (t *TupleConst) Equal(o Node) bool {
	if ot, ok := o.(*TupleConst); ok {
		if len(t.Vals) != len(ot.Vals) {
			return false
		}
		for i, v := range t.Vals {
			if !AstEqual(v, t.Vals[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (r *RecordConst) Equal(o Node) bool {
	if or, ok := o.(*RecordConst); ok {
		if len(r.Fields) != len(or.Fields) {
			return false
		}
		for key, val := range r.Fields {
			ov, ok := or.Fields[key]
			if !ok {
				return false
			}
			if !AstEqual(val, ov) {
				return false
			}
		}
		return true
	}
	return false
}

func (i *IfExpr) Equal(o Node) bool {
	if oi, ok := o.(*IfExpr); ok {
		return AstEqual(i.Cond, oi.Cond) && AstEqual(i.IfBranch, oi.IfBranch) && AstEqual(i.ElseBranch, oi.ElseBranch)
	}
	return false
}

func (w *WhileStmt) Equal(o Node) bool {
	if ow, ok := o.(*WhileStmt); ok {
		return AstEqual(w.Cond, ow.Cond) && AstEqual(w.Body, w.Body)
	}
	return false
}

func (l *LetExpr) Equal(o Node) bool {
	if ol, ok := o.(*LetExpr); ok {
		return AstEqual(l.Decls, ol.Decls) && AstEqual(l.Body, ol.Body)
	}
	return false
}

func (i *Identifier) Equal(o Node) bool {
	if oi, ok := o.(*Identifier); ok {
		return i.Name == oi.Name
	}
	return false
}

func (l *LambdaExpr) Equal(o Node) bool {
	if ol, ok := o.(*LambdaExpr); ok {
		if len(l.Args) != len(ol.Args) {
			return false
		}
		for i, arg := range l.Args {
			if arg.Name != ol.Args[i].Name {
				return false
			}
		}
		return AstEqual(l.Body, ol.Body)
	}
	return false
}

func (b *BoolConst) Equal(o Node) bool {
	if ob, ok := o.(*BoolConst); ok {
		return ob.Val == b.Val
	}
	return false
}

func (a *Assignment) Equal(o Node) bool {
	if oa, ok := o.(*Assignment); ok {
		return AstEqual(oa.LValue, a.LValue) && AstEqual(oa.RValue, a.RValue)
	}
	return false
}

func (n *NoneConst) Equal(o Node) bool {
	if _, ok := o.(*NoneConst); ok {
		return true
	}
	return false
}

func (a *Access) Equal(o Node) bool {
	if oa, ok := o.(*Access); ok {
		return AstEqual(a.Lhs, oa.Lhs) && a.Property.Name == oa.Property.Name
	}
	return false
}
