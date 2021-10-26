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
			"val a = 1",
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
			"\n\n\nval a = 1\n\n\n",
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
					Args: []ast.FuncDeclArg{},
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
					Args: []ast.FuncDeclArg{
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
							&ast.StmtExpr{&ast.IntConst{Val: 1}},
							&ast.StmtExpr{&ast.IntConst{Val: 2}},
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
							&ast.StmtExpr{&ast.IntConst{Val: 1}},
						},
					},
					ElseBranch: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{&ast.IntConst{Val: 2}},
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
							&ast.StmtExpr{&ast.IntConst{Val: 1}},
						},
					},
					ElseBranch: &ast.IfExpr{
						Cond: &ast.Identifier{Name: "b"},
						IfBranch: &ast.Block{
							Instr: []ast.Stmt{
								&ast.StmtExpr{&ast.IntConst{Val: 2}},
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
					Args: []ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{&ast.IfExpr{
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
										&ast.StmtExpr{&ast.IntConst{Val: 1}},
									},
								},
								ElseBranch: &ast.Block{
									Instr: []ast.Stmt{
										&ast.StmtExpr{&ast.IntConst{Val: 2}},
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
							&ast.StmtExpr{&ast.IntConst{Val: 1}},
							&ast.StmtExpr{&ast.IntConst{Val: 2}},
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
					Args: []ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{&ast.IntConst{Val: 1}},
							&ast.StmtExpr{&ast.IntConst{Val: 2}},
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
					Args: []ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{&ast.IntConst{Val: 1}},
							&ast.WhileStmt{
								Cond: &ast.Identifier{
									Name: "a",
								},
								Body: &ast.Block{
									Instr: []ast.Stmt{
										&ast.StmtExpr{&ast.IntConst{Val: 2}},
										&ast.StmtExpr{&ast.IntConst{Val: 3}},
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
					Args: []ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{&ast.IntConst{Val: 1}},
							&ast.WhileStmt{
								Cond: &ast.Identifier{
									Name: "a",
								},
								Body: &ast.Block{
									Instr: []ast.Stmt{
										&ast.StmtExpr{&ast.IntConst{Val: 2}},
										&ast.StmtExpr{&ast.IntConst{Val: 3}},
									},
								},
							},
							&ast.StmtExpr{&ast.IntConst{Val: 4}},
						},
					},
				},
			},
		},
	}
	matchAstWithTable(t, &table)
}

func TestTupleParsing(t *testing.T) {
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
	}

	matchAstWithTable(t, &table)
}

func TestParsingLambda(t *testing.T) {
	table := ptable{
		{
			"do -> a b",
			[]an{
				&ast.LambdaExpr{
					Args: []ast.FuncDeclArg{},
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
					Args: []ast.FuncDeclArg{
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
					Args: []ast.FuncDeclArg{},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{&ast.FuncApplication{
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
					Args: []ast.FuncDeclArg{
						{Name: "a"}, {Name: "b"}, {Name: "c"},
					},
					Body: &ast.Block{
						Instr: []ast.Stmt{
							&ast.StmtExpr{&ast.FuncApplication{
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
						Args: []ast.FuncDeclArg{},
						Body: &ast.Block{
							Instr: []ast.Stmt{
								&ast.StmtExpr{&ast.IntConst{Val: 1}},
								&ast.StmtExpr{&ast.IntConst{Val: 2}},
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
						Args: []ast.FuncDeclArg{},
						Body: &ast.Block{
							Instr: []ast.Stmt{
								&ast.StmtExpr{&ast.IntConst{Val: 1}},
								&ast.StmtExpr{&ast.IntConst{Val: 2}},
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
						Args: []ast.FuncDeclArg{},
						Body: &ast.Block{
							Instr: []ast.Stmt{
								&ast.StmtExpr{&ast.IntConst{Val: 1}},
								&ast.StmtExpr{&ast.IntConst{Val: 2}},
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
						Args: []ast.FuncDeclArg{},
						Body: &ast.Block{
							Instr: []ast.Stmt{
								&ast.StmtExpr{&ast.IntConst{Val: 1}},
								&ast.StmtExpr{&ast.IntConst{Val: 2}},
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
			"fn a:\n  val a = 2\n  val a = b\nval c = 1",
			[]an{
				&ast.FuncDecl{
					Name: "a",
					Args: []ast.FuncDeclArg{},
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
