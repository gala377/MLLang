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
	matchAllTestWithTable(t, &table)
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
	matchAllTestWithTable(t, &table)
}

func TestScanningIdent(t *testing.T) {
	table := tablet{
		{
			"ident1 __myident2?",
			[]it{
				{"ident1", token.Identifier, 0, 6},
				{"__myident2?", token.Identifier, 7, 18},
			},
		},
	}
	matchAllTestWithTable(t, &table)
}

func TestScanningNumbers(t *testing.T) {
	table := tablet{
		{
			"0 1 10",
			[]it{
				{"0", token.Integer, 0, 1},
				{"1", token.Integer, 2, 3},
				{"10", token.Integer, 4, 6},
			},
		},
		{
			"0.123 123.000 1234.1 1. 0.",
			[]it{
				{"0.123", token.Float, 0, 5},
				{"123.000", token.Float, 6, 13},
				{"1234.1", token.Float, 14, 20},
				{"1.", token.Float, 21, 23},
				{"0.", token.Float, 24, 26},
			},
		},
	}
	matchAllTestWithTable(t, &table)
}

func TestStringScanning(t *testing.T) {
	table := tablet{
		{
			"\"simple string\"",
			[]it{{"simple string", token.String, 0, 15}},
		},
		{
			"'simple string'",
			[]it{{"simple string", token.String, 0, 15}},
		},
		{
			"'simple\" \"string'",
			[]it{{"simple\" \"string", token.String, 0, 17}},
		},
		{
			"\"simple' 'string\"",
			[]it{{"simple' 'string", token.String, 0, 17}},
		},
		{
			"'a' 'b' 'c'",
			[]it{
				{"a", token.String, 0, 3},
				{"b", token.String, 4, 7},
				{"c", token.String, 8, 11},
			},
		},
	}
	matchAllTestWithTable(t, &table)
}

func matchAllTestWithTable(t *testing.T, table *tablet) {
	for _, test := range *table {
		t.Run(test.source, func(t *testing.T) {
			l := NewLexer(strings.NewReader(test.source), func(pos Position, msg string) {
				t.Fatalf("Error on pos %v with msg %v", pos, msg)
			})
			for _, want := range test.toks {
				got := l.Next()
				wanttok := token.Token{Typ: want.T, Val: want.N, Span: span.NewSpan(want.B, want.E)}
				if got != wanttok {
					t.Errorf("Wrong token - want: %#v got: %#v, lpos: %v", wanttok, got, l.position)
				}
			}
			eof := l.Next()
			if eof.Typ != token.Eof {
				t.Errorf("Expected EOF token, got: %v", eof)
			}
		})
	}
}

// struct { errorregex, pos }
// func matchErrorsWithTable
