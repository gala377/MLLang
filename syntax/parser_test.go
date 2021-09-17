package syntax

import (
	"log"
	"strings"
	"testing"

	"github.com/gala377/MLLang/syntax/ast"
	"github.com/gala377/MLLang/syntax/span"
)

type ptable []struct {
	source string
	expect []an
}

type an = ast.Node

var dummySpan span.Span = span.NewSpan(
	span.Position{Line: 0, Column: 0, Offset: 0},
	span.Position{Line: 0, Column: 0, Offset: 0},
)

func TestTopLevelValDecl(t *testing.T) {
	table := ptable{
		{
			"val a = 1",
			[]an{
				&ast.GlobalValDecl{
					Span: &dummySpan,
					Name: "a",
					Rhs: &ast.IntConst{
						Span: &dummySpan,
						Val:  1,
					},
				},
			},
		},
		{
			"\n\n\nval a = 1\n\n\n",
			[]an{
				&ast.GlobalValDecl{
					Span: &dummySpan,
					Name: "a",
					Rhs: &ast.IntConst{
						Span: &dummySpan,
						Val:  1,
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
					Span: &dummySpan,
					Name: "a",
				},
			},
		},
		{
			"1.123",
			[]an{
				&ast.FloatConst{
					Span: &dummySpan,
					Val:  1.123,
				},
			},
		},
		{
			"(a)",
			[]an{
				&ast.Identifier{
					Span: &dummySpan,
					Name: "a",
				},
			},
		},
		{
			"(1.123)",
			[]an{
				&ast.FloatConst{
					Span: &dummySpan,
					Val:  1.123,
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
					Span: &dummySpan,
					Callee: &ast.Identifier{
						Span: &dummySpan,
						Name: "a",
					},
					Args: []ast.Expr{
						&ast.Identifier{
							Span: &dummySpan,
							Name: "b",
						},
						&ast.Identifier{
							Span: &dummySpan,
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
					Span: &dummySpan,
					Callee: &ast.FuncApplication{
						Span: &dummySpan,
						Callee: &ast.Identifier{
							Span: &dummySpan,
							Name: "a",
						},
						Args: []ast.Expr{
							&ast.IntConst{
								Span: &dummySpan,
								Val:  1,
							},
						},
					},
					Args: []ast.Expr{
						&ast.Identifier{
							Span: &dummySpan,
							Name: "b",
						},
						&ast.FuncApplication{
							Span: &dummySpan,
							Callee: &ast.Identifier{
								Span: &dummySpan,
								Name: "c",
							},
							Args: []ast.Expr{
								&ast.IntConst{
									Span: &dummySpan,
									Val:  1,
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

func matchAstWithTable(t *testing.T, table *ptable) {
	for _, test := range *table {
		t.Run(test.source, func(t *testing.T) {
			log.Printf("SOURCE IS %v", test.source)
			p := NewParser(strings.NewReader(test.source))
			got := p.Parse()
			if len(got) != len(test.expect) {
				t.Fatalf("The parse result has wrong len\nwant: %v\ngot: %v", test.expect, got)
			}
			for i, want := range test.expect {
				if !ast.AstEqual(got[i], want) {
					t.Errorf("Mismatched node at position %v\nwant: %v\n got: %v", i, want, got[i])
				}
			}
			if len(p.errors) > 0 {
				t.Fatalf("Parser had unexpected errors %v", p.errors)
			}
		})
	}
}
