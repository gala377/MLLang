package syntax

import (
	"strings"
	"testing"
)

type it struct {
	N string
	T Id
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
			[]it{{"a", Identifier, 0, 1}, {"b", Identifier, 2, 3}, {"c", Identifier, 4, 5}},
		},
		{
			"a  b  c  ",
			[]it{{"a", Identifier, 0, 1}, {"b", Identifier, 3, 4}, {"c", Identifier, 6, 7}},
		},
	}
	for _, test := range table {
		t.Run(test.source, func(t *testing.T) {
			l := NewLexer(strings.NewReader(test.source))
			for _, want := range test.toks {
				got := l.Next()
				wanttok := Token{want.T, want.N, Span{want.B, want.E}}
				if got != wanttok {
					t.Errorf("Wrong token - want: %v got: %v", wanttok, got)
				}
			}
			eof := l.Next()
			if eof.typ != Eof {
				t.Errorf("Expected EOF token, got: %v", eof)
			}
		})
	}
}

func TestIndentScanning(t *testing.T) {
	table := tablet{
		{
			"    ",
			[]it{{"4", Indent, 0, 4}},
		},
		{
			"     \n     \n     ",
			[]it{
				{"5", Indent, 0, 5},
				{"\n", NewLine, 5, 6},
				{"5", Indent, 6, 11},
				{"\n", NewLine, 11, 12},
				{"5", Indent, 12, 17},
			},
		},
		{
			"aa\r\n     \n \n\n",
			[]it{
				{"aa", Identifier, 0, 2},
				{"\n", NewLine, 3, 4},
				{"5", Indent, 4, 9},
				{"\n", NewLine, 9, 10},
				{"1", Indent, 10, 11},
				{"\n", NewLine, 11, 12},
				{"\n", NewLine, 12, 13},
			},
		},
	}
	for _, test := range table {
		l := NewLexer(strings.NewReader(test.source))
		for _, want := range test.toks {
			got := l.Next()
			wanttok := Token{want.T, want.N, Span{want.B, want.E}}
			if got != wanttok {
				t.Errorf("Source: %#v - Wrong token - want: %#v got: %#v, lpos: %v", test.source, wanttok, got, l.position)
			}
		}
		eof := l.Next()
		if eof.typ != Eof {
			t.Errorf("Expected EOF token, got: %v", eof)
		}
	}
}

func TestScanningIdent(t *testing.T) {
	source := "ident1 __myident2"
	l := NewLexer(strings.NewReader(source))
	got := l.Next()
	want := Token{Identifier, "ident1", Span{0, 6}}
	if got != want {
		t.Errorf("Wrong current token - want: %v got: %v", want, got)
	}
	got = l.Peek()
	want = Token{Identifier, "__myident2", Span{7, 17}}
	if got != want {
		t.Errorf("Wrong peek token - want: %v got: %v", want, got)
	}
}
