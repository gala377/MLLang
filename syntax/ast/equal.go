package ast

type NodeEqual interface {
	Node
	Equal(Node) bool
}

func AstEqual(n1 Node, n2 Node) bool {
	cmp, ok := n1.(NodeEqual)
	if !ok {
		// msg := fmt.Sprintf("Comparision unsupported for node: %#v", n1)
		// panic(msg)
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
		return AstEqual(&f.Body, &of.Body)
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
		return true
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
		return AstEqual(i.Cond, oi.Cond) && AstEqual(&i.IfBranch, &oi.IfBranch) && AstEqual(&i.ElseBranch, &oi.ElseBranch)
	}
	return false
}

func (w *WhileExpr) Equal(o Node) bool {
	if ow, ok := o.(*WhileExpr); ok {
		return AstEqual(w.Cond, ow.Cond) && AstEqual(&w.Body, &ow.Body)
	}
	return false
}

func (l *LetExpr) Equal(o Node) bool {
	if ol, ok := o.(*LetExpr); ok {
		return AstEqual(l.Decls, ol.Decls) && AstEqual(&l.Body, &ol.Body)
	}
	return false
}

func (i *Identifier) Equal(o Node) bool {
	if oi, ok := o.(*Identifier); ok {
		return i.Name == oi.Name
	}
	return false
}
