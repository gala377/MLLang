package syntax

import (
	"strings"
	"testing"

	"github.com/gala377/MLLang/syntax/ast"
)

type ptable []struct {
	source string
	expect []an
}

type an = ast.Node

func TestTopLevelValDecl(t *testing.T) {
	table := ptable{
		{
			"let a = 1",
			[]an{
				&ast.GlobalValDecl{
					Name: "a",
					Rhs: &ast.IntConst{
						Val: 1,
					},
				},
			},
		},
		{
			"\n\n\nlet a = 1\n\n\n",
			[]an{
				&ast.GlobalValDecl{
					Name: "a",
					Rhs: &ast.IntConst{
						Val: 1,
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestPrimaryExpressions(t *testing.T) {
	table := ptable{
		{
			"a",
			[]an{
				&ast.Identifier{
					Name: "a",
				},
			},
		},
		{
			"1.123",
			[]an{
				&ast.FloatConst{
					Val: 1.123,
				},
			},
		},
		{
			"(a)",
			[]an{
				&ast.Identifier{
					Name: "a",
				},
			},
		},
		{
			"(1.123)",
			[]an{
				&ast.FloatConst{
					Val: 1.123,
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestFuncApplication(t *testing.T) {
	table := ptable{
		{
			"a b c",
			[]an{
				&ast.FuncApplication{
					Callee: &ast.Identifier{
						Name: "a",
					},
					Args: []ast.Expr{
						&ast.Identifier{
							Name: "b",
						},
						&ast.Identifier{
							Name: "c",
						},
					},
				},
			},
		},
		{
			"(a 1) b (c 1)",
			[]an{
				&ast.FuncApplication{
					Callee: &ast.FuncApplication{
						Callee: &ast.Identifier{
							Name: "a",
						},
						Args: []ast.Expr{
							&ast.IntConst{
								Val: 1,
							},
						},
					},
					Args: []ast.Expr{
						&ast.Identifier{
							Name: "b",
						},
						&ast.FuncApplication{
							Callee: &ast.Identifier{
								Name: "c",
							},
							Args: []ast.Expr{
								&ast.IntConst{
									Val: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			"((a 1)!)!",
			[]an{
				&ast.FuncApplication{
					Callee: &ast.FuncApplication{
						Callee: &ast.FuncApplication{
							Callee: &ast.Identifier{
								Name: "a",
							},
							Args: []ast.Expr{
								&ast.IntConst{
									Val: 1,
								},
							},
						},
						Args: []ast.Expr{},
					},
					Args: []ast.Expr{},
				},
			},
		},
		{
			"a!",
			[]an{
				&ast.FuncApplication{
					Callee: &ast.Identifier{
						Name: "a",
					},
					Args: []ast.Expr{},
				},
			},
		},
		{
			"a! b! c",
			[]an{
				&ast.FuncApplication{
					Callee: &ast.FuncApplication{
						Callee: &ast.Identifier{Name: "a"},
						Args:   make([]ast.Expr, 0),
					},
					Args: []ast.Expr{
						&ast.FuncApplication{
							Callee: &ast.Identifier{Name: "b"},
							Args:   make([]ast.Expr, 0),
						},
						&ast.Identifier{Name: "c"},
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestFuncDeclaration(t *testing.T) {
	table := ptable{
		{
			"fn a = 1",
			[]an{
				&ast.FuncDecl{
					Name: "a",
					Args: []*ast.FuncDeclArg{},
					Body: &ast.IntConst{
						Val: 1,
					},
				},
			},
		},
		{
			"fn a b c = (b)",
			[]an{
				&ast.FuncDecl{
					Name: "a",
					Args: []*ast.FuncDeclArg{
						{Name: "b"},
						{Name: "c"},
					},
					Body: &ast.Identifier{
						Name: "b",
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestParsingIf(t *testing.T) {
	table := ptable{
		{
			"if a 1:\n" +
				"  1\n" +
				"  2",
			[]an{
				&ast.IfExpr{
					Cond: &ast.FuncApplication{
						Callee: &ast.Identifier{
							Name: "a",
						},
						Args: []ast.Expr{
							&ast.IntConst{Val: 1},
						},
					},
					IfBranch: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
						},
					},
					ElseBranch: nil,
				},
			},
		},
		{
			"if a 1:\n" +
				"  1\n" +
				"else:\n" +
				"  2",
			[]an{
				&ast.IfExpr{
					Cond: &ast.FuncApplication{
						Callee: &ast.Identifier{
							Name: "a",
						},
						Args: []ast.Expr{
							&ast.IntConst{Val: 1},
						},
					},
					IfBranch: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
						},
					},
					ElseBranch: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
						},
					},
				},
			},
		},
		{
			"if a 1:\n" +
				"  1\n" +
				"else if b:\n" +
				"  2",
			[]an{
				&ast.IfExpr{
					Cond: &ast.FuncApplication{
						Callee: &ast.Identifier{
							Name: "a",
						},
						Args: []ast.Expr{
							&ast.IntConst{Val: 1},
						},
					},
					IfBranch: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
						},
					},
					ElseBranch: &ast.IfExpr{
						Cond: &ast.Identifier{Name: "b"},
						IfBranch: &ast.Block{
							Instr: []ast.Stmt{
								&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
							},
						},
						ElseBranch: nil,
					},
				},
			},
		},
		{
			"fn a:\n" +
				"  if a 1:\n" +
				"    1\n" +
				"  else:\n" +
				"    2",
			[]an{
				&ast.FuncDecl{
					Name: "a",
					Args: []*ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IfExpr{
								Cond: &ast.FuncApplication{
									Callee: &ast.Identifier{
										Name: "a",
									},
									Args: []ast.Expr{
										&ast.IntConst{Val: 1},
									},
								},
								IfBranch: &ast.Block{
									Instr: []ast.Stmt{
										&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
									},
								},
								ElseBranch: &ast.Block{
									Instr: []ast.Stmt{
										&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
									},
								},
							}},
						},
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestParsingWhile(t *testing.T) {
	table := ptable{
		{
			"while (a 1):\n" +
				"  1\n" +
				"  2",
			[]an{
				&ast.WhileStmt{
					Cond: &ast.FuncApplication{
						Callee: &ast.Identifier{
							Name: "a",
						},
						Args: []ast.Expr{
							&ast.IntConst{Val: 1},
						},
					},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
						},
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestParsingBlocks(t *testing.T) {
	table := ptable{
		{
			"fn a:\n" +
				"  1\n" +
				"  2",
			[]an{
				&ast.FuncDecl{
					Name: "a",
					Args: []*ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
						},
					},
				},
			},
		},
		{
			"fn a:\n" +
				"  1\n" +
				"  while a:\n" +
				"    2\n" +
				"    3\n",
			[]an{
				&ast.FuncDecl{
					Name: "a",
					Args: []*ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
							&ast.WhileStmt{
								Cond: &ast.Identifier{
									Name: "a",
								},
								Body: &ast.Block{
									Instr: []ast.Stmt{
										&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
										&ast.StmtExpr{Expr: &ast.IntConst{Val: 3}},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"fn a:\n" +
				"  1\n" +
				"  while(a):\n" +
				"    2\n" +
				"    3\n" +
				"  4",
			[]an{
				&ast.FuncDecl{
					Name: "a",
					Args: []*ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
							&ast.WhileStmt{
								Cond: &ast.Identifier{
									Name: "a",
								},
								Body: &ast.Block{
									Instr: []ast.Stmt{
										&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
										&ast.StmtExpr{Expr: &ast.IntConst{Val: 3}},
									},
								},
							},
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 4}},
						},
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestTupleSequences(t *testing.T) {
	table := ptable{
		{
			"()",
			[]an{
				&ast.TupleConst{
					Vals: []ast.Expr{},
				},
			},
		},
		{
			"(a)",
			[]an{
				&ast.Identifier{
					Name: "a",
				},
			},
		},
		{
			"(a,)",
			[]an{
				&ast.TupleConst{
					Vals: []ast.Expr{
						&ast.Identifier{
							Name: "a",
						},
					},
				},
			},
		},
		{
			"(a, a b, 3)",
			[]an{
				&ast.TupleConst{
					Vals: []ast.Expr{
						&ast.Identifier{
							Name: "a",
						},
						&ast.FuncApplication{
							Callee: &ast.Identifier{
								Name: "a",
							},
							Args: []ast.Expr{
								&ast.Identifier{
									Name: "b",
								},
							},
						},
						&ast.IntConst{
							Val: 3,
						},
					},
				},
			},
		},
		{
			"(a, a b, 3,)",
			[]an{
				&ast.TupleConst{
					Vals: []ast.Expr{
						&ast.Identifier{
							Name: "a",
						},
						&ast.FuncApplication{
							Callee: &ast.Identifier{
								Name: "a",
							},
							Args: []ast.Expr{
								&ast.Identifier{
									Name: "b",
								},
							},
						},
						&ast.IntConst{
							Val: 3,
						},
					},
				},
			},
		},
		{
			"[]",
			[]an{
				&ast.ListConst{
					Vals: []ast.Expr{},
				},
			},
		},
		{
			"[ a ]",
			[]an{
				&ast.ListConst{
					Vals: []ast.Expr{
						&ast.Identifier{
							Name: "a",
						},
					},
				},
			},
		},
		{
			"[ a, ]",
			[]an{
				&ast.ListConst{
					Vals: []ast.Expr{
						&ast.Identifier{
							Name: "a",
						},
					},
				},
			},
		},
		{
			"[a, a b, 3]",
			[]an{
				&ast.ListConst{
					Vals: []ast.Expr{
						&ast.Identifier{
							Name: "a",
						},
						&ast.FuncApplication{
							Callee: &ast.Identifier{
								Name: "a",
							},
							Args: []ast.Expr{
								&ast.Identifier{
									Name: "b",
								},
							},
						},
						&ast.IntConst{
							Val: 3,
						},
					},
				},
			},
		},
		{
			"[ a, a b, 3, ]",
			[]an{
				&ast.ListConst{
					Vals: []ast.Expr{
						&ast.Identifier{
							Name: "a",
						},
						&ast.FuncApplication{
							Callee: &ast.Identifier{
								Name: "a",
							},
							Args: []ast.Expr{
								&ast.Identifier{
									Name: "b",
								},
							},
						},
						&ast.IntConst{
							Val: 3,
						},
					},
				},
			},
		},
		{
			"[1, 2,\n 3, 4,\n 5]",
			[]an{
				&ast.ListConst{
					Vals: []ast.Expr{
						&ast.IntConst{Val: 1},
						&ast.IntConst{Val: 2},
						&ast.IntConst{Val: 3},
						&ast.IntConst{Val: 4},
						&ast.IntConst{Val: 5},
					},
				},
			},
		},
		{
			"[1, 2,\n 3, 4,\n 5,\n]",
			[]an{
				&ast.ListConst{
					Vals: []ast.Expr{
						&ast.IntConst{Val: 1},
						&ast.IntConst{Val: 2},
						&ast.IntConst{Val: 3},
						&ast.IntConst{Val: 4},
						&ast.IntConst{Val: 5},
					},
				},
			},
		},
	}

	matchAstWithTable(t, &table)
}

func TestParsingLambda(t *testing.T) {
	table := ptable{
		{
			"do -> a b",
			[]an{
				&ast.LambdaExpr{
					Args: []*ast.FuncDeclArg{},
					Body: &ast.FuncApplication{
						Callee: &ast.Identifier{Name: "a"},
						Args: []ast.Expr{
							&ast.Identifier{Name: "b"},
						},
					},
				},
			},
		},
		{
			"do |a b c| -> a b",
			[]an{
				&ast.LambdaExpr{
					Args: []*ast.FuncDeclArg{
						{Name: "a"}, {Name: "b"}, {Name: "c"},
					},
					Body: &ast.FuncApplication{
						Callee: &ast.Identifier{Name: "a"},
						Args: []ast.Expr{
							&ast.Identifier{Name: "b"},
						},
					},
				},
			},
		},
		{
			"do:\n  a b",
			[]an{
				&ast.LambdaExpr{
					Args: []*ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.FuncApplication{
								Callee: &ast.Identifier{Name: "a"},
								Args: []ast.Expr{
									&ast.Identifier{Name: "b"},
								},
							}},
						},
					},
				},
			},
		},

		{
			"do |a b c|:\n  a b",
			[]an{
				&ast.LambdaExpr{
					Args: []*ast.FuncDeclArg{
						{Name: "a"}, {Name: "b"}, {Name: "c"},
					},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.FuncApplication{
								Callee: &ast.Identifier{Name: "a"},
								Args: []ast.Expr{
									&ast.Identifier{Name: "b"},
								},
							}},
						},
					},
				},
			},
		},
		{
			"do a b c -> a b",
			[]an{
				&ast.LambdaExpr{
					Args: []*ast.FuncDeclArg{
						{Name: "a"}, {Name: "b"}, {Name: "c"},
					},
					Body: &ast.FuncApplication{
						Callee: &ast.Identifier{Name: "a"},
						Args: []ast.Expr{
							&ast.Identifier{Name: "b"},
						},
					},
				},
			},
		},
		{
			"do a b c:\n  a b",
			[]an{
				&ast.LambdaExpr{
					Args: []*ast.FuncDeclArg{
						{Name: "a"}, {Name: "b"}, {Name: "c"},
					},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.FuncApplication{
								Callee: &ast.Identifier{Name: "a"},
								Args: []ast.Expr{
									&ast.Identifier{Name: "b"},
								},
							}},
						},
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestFuncBlockApplication(t *testing.T) {
	table := ptable{
		{
			"a:\n  1\n  2",
			[]an{
				&ast.FuncApplication{
					Callee: &ast.Identifier{
						Name: "a",
					},
					Args: []ast.Expr{},
					Block: &ast.LambdaExpr{
						Args: []*ast.FuncDeclArg{},
						Body: &ast.Block{
							Instr: []ast.Stmt{
								&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
								&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
							},
						},
					},
				},
			},
		},
		{
			"a b (c):\n  1\n  2",
			[]an{
				&ast.FuncApplication{
					Callee: &ast.Identifier{
						Name: "a",
					},
					Args: []ast.Expr{&ast.Identifier{Name: "b"}, &ast.Identifier{Name: "c"}},
					Block: &ast.LambdaExpr{
						Args: []*ast.FuncDeclArg{},
						Body: &ast.Block{
							Instr: []ast.Stmt{
								&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
								&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
							},
						},
					},
				},
			},
		},
		{
			"a b (c 6):\n  1\n  2",
			[]an{
				&ast.FuncApplication{
					Callee: &ast.Identifier{
						Name: "a",
					},
					Args: []ast.Expr{
						&ast.Identifier{Name: "b"},
						&ast.FuncApplication{
							Callee: &ast.Identifier{Name: "c"},
							Args:   []ast.Expr{&ast.IntConst{Val: 6}},
						},
					},
					Block: &ast.LambdaExpr{
						Args: []*ast.FuncDeclArg{},
						Body: &ast.Block{
							Instr: []ast.Stmt{
								&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
								&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
							},
						},
					},
				},
			},
		},
		{
			"(a 1 2):\n  1\n  2",
			[]an{
				&ast.FuncApplication{
					Callee: &ast.FuncApplication{
						Callee: &ast.Identifier{Name: "a"},
						Args: []ast.Expr{
							&ast.IntConst{Val: 1},
							&ast.IntConst{Val: 2},
						},
					},
					Args: []ast.Expr{},
					Block: &ast.LambdaExpr{
						Args: []*ast.FuncDeclArg{},
						Body: &ast.Block{
							Instr: []ast.Stmt{
								&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
								&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
							},
						},
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestLocalVariableDecl(t *testing.T) {
	table := ptable{
		{
			"fn a:\n  let a = 2\n  let a = b\nlet c = 1",
			[]an{
				&ast.FuncDecl{
					Name: "a",
					Args: []*ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.ValDecl{
								Name: "a",
								Rhs:  &ast.IntConst{Val: 2},
							},
							&ast.ValDecl{
								Name: "a",
								Rhs:  &ast.Identifier{Name: "b"}},
						},
					},
				},
				&ast.GlobalValDecl{
					Name: "c",
					Rhs:  &ast.IntConst{Val: 1},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestAssignment(t *testing.T) {
	table := ptable{
		{
			"a = 1\na b c = d a\n",
			[]an{
				&ast.Assignment{
					LValue: &ast.Identifier{Name: "a"},
					RValue: &ast.IntConst{Val: 1},
				},
				&ast.Assignment{
					LValue: &ast.FuncApplication{
						Callee: &ast.Identifier{Name: "a"},
						Args: []ast.Expr{
							&ast.Identifier{Name: "b"},
							&ast.Identifier{Name: "c"},
						},
					},
					RValue: &ast.FuncApplication{
						Callee: &ast.Identifier{Name: "d"},
						Args: []ast.Expr{
							&ast.Identifier{Name: "a"},
						},
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestRecordLiteral(t *testing.T) {
	table := ptable{
		{
			"{ a: a b, b : 1 }",
			[]an{
				&ast.RecordConst{
					Fields: []ast.RecordField{
						{Key: "a", Val: &ast.FuncApplication{
							Callee: &ast.Identifier{Name: "a"},
							Args: []ast.Expr{
								&ast.Identifier{Name: "b"},
							},
						}},
						{Key: "b", Val: &ast.IntConst{Val: 1}},
					},
				},
			},
		},
		{
			"{a: {a: {a: 1}}}",
			[]an{
				&ast.RecordConst{
					Fields: []ast.RecordField{
						{Key: "a", Val: &ast.RecordConst{
							Fields: []ast.RecordField{
								{Key: "a", Val: &ast.RecordConst{
									Fields: []ast.RecordField{
										{Key: "a", Val: &ast.IntConst{Val: 1}},
									},
								}},
							},
						}},
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestParsingAccess(t *testing.T) {
	table := ptable{
		{
			"a.b",
			[]an{
				&ast.Access{
					Lhs:      &ast.Identifier{Name: "a"},
					Property: ast.Identifier{Name: "b"},
				},
			},
		},
		{
			"(1, 2).hello",
			[]an{
				&ast.Access{
					Lhs: &ast.TupleConst{
						Vals: []ast.Expr{&ast.IntConst{Val: 1}, &ast.IntConst{Val: 2}},
					},
					Property: ast.Identifier{Name: "hello"},
				},
			},
		},
		{
			"a.b.c!.d.e!.f",
			[]an{
				&ast.Access{
					Lhs: &ast.FuncApplication{
						Callee: &ast.Access{
							Lhs: &ast.Access{
								Lhs: &ast.FuncApplication{
									Callee: &ast.Access{
										Lhs: &ast.Access{
											Lhs:      &ast.Identifier{Name: "a"},
											Property: ast.Identifier{Name: "b"},
										},
										Property: ast.Identifier{Name: "c"},
									},
									Args: []ast.Expr{},
								},
								Property: ast.Identifier{Name: "d"},
							},
							Property: ast.Identifier{Name: "e"},
						},
						Args: []ast.Expr{},
					},
					Property: ast.Identifier{Name: "f"},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestParsingHandle(t *testing.T) {
	table := ptable{
		{
			"handle:\n 1\nwith e1 a1 -> k:\n 2\nwith e2 a2:\n 3\n",
			[]an{
				&ast.Handle{
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
						},
					},
					Arms: []*ast.WithClause{
						{
							Effect:       &ast.Identifier{Name: "e1"},
							Arg:          &ast.FuncDeclArg{Name: "a1"},
							Continuation: &ast.FuncDeclArg{Name: "k"},
							Body: &ast.Block{
								Instr: []ast.Stmt{
									&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
								},
							},
						},
						{
							Effect:       &ast.Identifier{Name: "e2"},
							Arg:          &ast.FuncDeclArg{Name: "a2"},
							Continuation: nil,
							Body: &ast.Block{
								Instr: []ast.Stmt{
									&ast.StmtExpr{Expr: &ast.IntConst{Val: 3}},
								},
							},
						},
					},
				},
			},
		},
		{
			"handle:\n 1\nwith e a -> k:\n 2",
			[]an{
				&ast.Handle{
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
						},
					},
					Arms: []*ast.WithClause{
						{
							Effect:       &ast.Identifier{Name: "e"},
							Arg:          &ast.FuncDeclArg{Name: "a"},
							Continuation: &ast.FuncDeclArg{Name: "k"},
							Body: &ast.Block{
								Instr: []ast.Stmt{
									&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
								},
							},
						},
					},
				},
			},
		},
		{
			"handle:\n 1\nwith e a if eq? a:\n 2",
			[]an{
				&ast.Handle{
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{Expr: &ast.IntConst{Val: 1}},
						},
					},
					Arms: []*ast.WithClause{
						{
							Effect: &ast.Identifier{Name: "e"},
							Arg:    &ast.FuncDeclArg{Name: "a"},
							Guard: &ast.FuncApplication{
								Callee: &ast.Identifier{Name: "eq?"},
								Args: []ast.Expr{
									&ast.Identifier{Name: "a"},
								},
							},
							Body: &ast.Block{
								Instr: []ast.Stmt{
									&ast.StmtExpr{Expr: &ast.IntConst{Val: 2}},
								},
							},
						},
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestParsingEffects(t *testing.T) {
	table := ptable{
		{
			"effect a\n",
			[]an{
				&ast.EffectDecl{
					Name: "a",
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func matchAstWithTable(t *testing.T, table *ptable) {
	for _, test := range *table {
		t.Run(test.source, func(t *testing.T) {
			t.Logf("SOURCE IS %v", test.source)
			p := NewParser(strings.NewReader(test.source))
			got := p.Parse()
			if len(got) != len(test.expect) {
				t.Logf("Parser errors %v", p.errors)
				t.Fatalf("The parse result has wrong len\nwant: %v\ngot: %v", test.expect, got)
			}
			for i, want := range test.expect {
				wantstmt := want
				if v, ok := want.(ast.Expr); ok {
					wantstmt = &ast.StmtExpr{Expr: v}
				}
				if !ast.AstEqual(got[i], wantstmt) {
					t.Errorf("Mismatched node at position %v\nwant: %v\n got: %v", i, want, got[i])
				}
			}
			if len(p.errors) > 0 {
				t.Fatalf("Parser had unexpected errors %v", p.errors)
			}
		})
	}
}
