package ast

type Visitor interface {
	Visit(n Node) Visitor
}

func Walk(v Visitor, n Node) {
	if v = v.Visit(n); v == nil {
		return
	}
	switch n := n.(type) {
	case *GlobalValDecl:
		Walk(v, n.Rhs)
	case *FuncDecl:
		Walk(v, n.Body)
	case *Block:
		for _, e := range n.Instr {
			Walk(v, e)
		}
	case *FuncApplication:
		Walk(v, n.Callee)
		for _, arg := range n.Args {
			Walk(v, arg)
		}
	case *RecordConst:
		for _, val := range n.Fields {
			Walk(v, val)
		}
	case *ListConst:
		for _, val := range n.Vals {
			Walk(v, val)
		}
	case *TupleConst:
		for _, val := range n.Vals {
			Walk(v, val)
		}
	case *IfExpr:
		Walk(v, n.Cond)
		Walk(v, n.IfBranch)
		Walk(v, n.ElseBranch)
	case *WhileStmt:
		Walk(v, n.Cond)
		Walk(v, n.Body)
	case *LetExpr:
		Walk(v, n.Decls)
		Walk(v, n.Body)
	case *LambdaExpr:
		Walk(v, n.Body)
	}
}
