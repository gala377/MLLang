package codegen

import (
	"fmt"
	"strings"

	"github.com/gala377/MLLang/syntax/ast"
)

var ARG_NAME = "@arg"
var CONT_NAME = "@cont"
var EFFECT_PREFIX = "@effect@"

type ClausGroup struct {
	Effect  ast.Expr
	Clauses []*ast.WithClause
}

func desugarGuards(h *ast.Handle, effectc int) (*ast.Handle, []*ast.ValDecl) {
	res := ast.Handle{
		Span: h.Span,
		Body: h.Body,
	}
	arms := groupClauses(h.Arms)
	savedeff := make([]*ast.ValDecl, 0, len(arms))
	harms := make([]*ast.WithClause, 0, len(arms))
	for path, cg := range arms {
		effect := fmt.Sprint(effectc) + EFFECT_PREFIX + path
		harms = append(harms, mergeClauseGroup(effect, h, cg))
		savedeff = append(savedeff, &ast.ValDecl{
			Span: cg.Effect.NodeSpan(),
			Name: effect,
			Rhs:  cg.Effect,
			Lift: true,
		})
	}
	res.Arms = harms
	return &res, savedeff
}

func groupClauses(harms []*ast.WithClause) map[string]*ClausGroup {
	arms := make(map[string]*ClausGroup)
	for _, arm := range harms {
		key := pathToStringReversed(arm.Effect)
		if _, ok := arms[key]; ok {
			g := arms[key]
			g.Clauses = append(g.Clauses, arm)
		} else {
			arms[key] = &ClausGroup{
				Effect:  arm.Effect,
				Clauses: []*ast.WithClause{arm},
			}
		}
	}
	return arms
}

func mergeClauseGroup(effect string, h *ast.Handle, cg *ClausGroup) *ast.WithClause {
	body := ast.Block{
		Span:  h.Span,
		Instr: []ast.Stmt{},
	}
	var catchall *ast.WithClause = nil
	var test ast.Expr = nil
	var currt ast.Expr = nil
	for _, c := range cg.Clauses {
		if c.Guard == nil {
			if catchall != nil {
				panic("two catch all with clauses for the same effect")
			}
			catchall = c
			continue
		}
		if test == nil {
			test = addGuardTest(test, c)
			currt = test
			continue
		}
		currt = addGuardTest(currt, c)
	}
	if test != nil {
		addCatchAll(effect, currt.(*ast.IfExpr), catchall)
		body.Instr = append(body.Instr, &ast.StmtExpr{Expr: test})
	} else {
		return catchall
	}
	return &ast.WithClause{
		Span:         h.Span,
		Effect:       &ast.Identifier{Span: h.Span, Name: effect},
		Arg:          &ast.FuncDeclArg{Span: h.Span, Name: ARG_NAME},
		Continuation: &ast.FuncDeclArg{Span: h.Span, Name: CONT_NAME},
		Body:         &body,
	}
}

func addCatchAll(effect string, testt *ast.IfExpr, catchall *ast.WithClause) {
	if catchall == nil {
		b := &ast.Block{
			Span: testt.Span,
			Instr: []ast.Stmt{
				&ast.StmtExpr{
					Expr: &ast.Resume{
						Span: testt.Span,
						Cont: &ast.Identifier{
							Span: testt.Span,
							Name: CONT_NAME,
						},
						Arg: &ast.FuncApplication{
							Span: testt.Span,
							Callee: &ast.Identifier{
								Span: testt.Span,
								Name: effect,
							},
							Args: []ast.Expr{
								&ast.Identifier{
									Span: testt.Span,
									Name: ARG_NAME,
								},
							},
						},
					},
				},
			},
		}
		testt.ElseBranch = b
	} else {
		testt.ElseBranch = createDesugaredBodyForWithClause(catchall)
	}
}

func addGuardTest(test ast.Expr, wc *ast.WithClause) ast.Expr {
	if test == nil {
		return &ast.IfExpr{
			Span:       wc.Span,
			Cond:       desugarGuardExpr(wc),
			IfBranch:   createDesugaredBodyForWithClause(wc),
			ElseBranch: nil,
		}
	}
	asif := test.(*ast.IfExpr)
	asif.ElseBranch = addGuardTest(asif.ElseBranch, wc)
	return asif.ElseBranch
}

func desugarGuardExpr(wc *ast.WithClause) ast.Expr {
	return &ast.FuncApplication{
		Span:   wc.Guard.NodeSpan(),
		Callee: wc.Guard,
		Args:   []ast.Expr{&ast.Identifier{Span: wc.Guard.NodeSpan(), Name: ARG_NAME}},
	}
}

func createDesugaredBodyForWithClause(wc *ast.WithClause) *ast.Block {
	args := make([]*ast.FuncDeclArg, 0, 2)
	callargs := make([]ast.Expr, 0, 2)
	args = append(args, wc.Arg)
	callargs = append(callargs, &ast.Identifier{
		Span: wc.Arg.Span,
		Name: ARG_NAME,
	})
	if wc.Continuation != nil {
		args = append(args, wc.Continuation)
		callargs = append(callargs, &ast.Identifier{
			Span: wc.Continuation.Span,
			Name: CONT_NAME,
		})
	}
	fun := ast.LambdaExpr{
		Span: wc.Body.Span,
		Args: args,
		Body: wc.Body,
	}
	return &ast.Block{
		Span: wc.Span,
		Instr: []ast.Stmt{
			&ast.StmtExpr{Expr: &ast.FuncApplication{
				Span:   wc.Span,
				Callee: &fun,
				Args:   callargs,
				Block:  nil,
			}},
		},
	}
}

func pathToStringReversed(p ast.Expr) string {
	var b strings.Builder
	for {
		switch v := p.(type) {
		case *ast.Access:
			b.WriteString(v.Property.Name)
			b.WriteByte('.')
			p = v.Lhs
		case *ast.Identifier:
			b.WriteString(v.Name)
			return b.String()
		default:
			panic("ICE: unexpected node in the path")
		}
	}
}
