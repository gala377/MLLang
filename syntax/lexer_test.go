package syntax

import (
	"strings"
	"testing"

	"github.com/gala377/MLLang/syntax/span"
	"github.com/gala377/MLLang/syntax/token"
)

type it struct {
	N string
	T token.Id
	B uint
	E uint
}

type tablet []struct {
	source string
	toks   []it
}

func TestScanningOneLetterIdent(t *testing.T) {
	table := tablet{
		{
			"a b c",
			[]it{{"a", token.Identifier, 0, 1}, {"b", token.Identifier, 2, 3}, {"c", token.Identifier, 4, 5}},
		},
		{
			"a  b  c  ",
			[]it{{"a", token.Identifier, 0, 1}, {"b", token.Identifier, 3, 4}, {"c", token.Identifier, 6, 7}},
		},
	}
	for _, test := range table {
		t.Run(test.source, func(t *testing.T) {
			l := NewLexer(strings.NewReader(test.source))
			for _, want := range test.toks {
				got := l.Next()
				wanttok := token.Token{want.T, want.N, span.Span{want.B, want.E}}
				if got != wanttok {
					t.Errorf("Wrong token - want: %v got: %v", wanttok, got)
				}
			}
			eof := l.Next()
			if eof.Typ != token.Eof {
				t.Errorf("Expected EOF token, got: %v", eof)
			}
		})
	}
}

func TestIndentScanning(t *testing.T) {
	table := tablet{
		{
			"    ",
			[]it{{"4", token.Indent, 0, 4}},
		},
		{
			"     \n     \n     ",
			[]it{
				{"5", token.Indent, 0, 5},
				{"\n", token.NewLine, 5, 6},
				{"5", token.Indent, 6, 11},
				{"\n", token.NewLine, 11, 12},
				{"5", token.Indent, 12, 17},
			},
		},
		{
			"aa\r\n     \n \n\n",
			[]it{
				{"aa", token.Identifier, 0, 2},
				{"\n", token.NewLine, 3, 4},
				{"5", token.Indent, 4, 9},
				{"\n", token.NewLine, 9, 10},
				{"1", token.Indent, 10, 11},
				{"\n", token.NewLine, 11, 12},
				{"\n", token.NewLine, 12, 13},
			},
		},
	}
	for _, test := range table {
		l := NewLexer(strings.NewReader(test.source))
		for _, want := range test.toks {
			got := l.Next()
			wanttok := token.Token{want.T, want.N, span.Span{want.B, want.E}}
			if got != wanttok {
				t.Errorf("Source: %#v - Wrong token - want: %#v got: %#v, lpos: %v", test.source, wanttok, got, l.position)
			}
		}
		eof := l.Next()
		if eof.Typ != token.Eof {
			t.Errorf("Expected EOF token, got: %v", eof)
		}
	}
}

func TestScanningIdent(t *testing.T) {
	source := "ident1 __myident2"
	l := NewLexer(strings.NewReader(source))
	got := l.Next()
	want := token.Token{token.Identifier, "ident1", span.Span{0, 6}}
	if got != want {
		t.Errorf("Wrong current token - want: %v got: %v", want, got)
	}
	got = l.Peek()
	want = token.Token{token.Identifier, "__myident2", span.Span{7, 17}}
	if got != want {
		t.Errorf("Wrong peek token - want: %v got: %v", want, got)
	}
}
