package syntax

import (
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
	}
	matchAstWithTable(t, &table)
}

func matchAstWithTable(t *testing.T, table *ptable) {
	for _, test := range *table {
		t.Run(test.source, func(t *testing.T) {
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
		})
	}
}
