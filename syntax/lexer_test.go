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
		{
			"fn val if else while match let macro",
			[]it{
				{"fn", token.Fn, 0, 2},
				{"val", token.Val, 3, 6},
				{"if", token.If, 7, 9},
				{"else", token.Else, 10, 14},
				{"while", token.While, 15, 20},
				{"match", token.Match, 21, 26},
				{"let", token.Let, 27, 30},
				{"macro", token.Macro, 31, 36},
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

func TestScanningOperators(t *testing.T) {
	table := tablet{
		{
			"=",
			[]it{
				{"=", token.Assignment, 0, 1},
			},
		},
		{
			"- ==> <> <$> %^ *&^% ;|- ??",
			[]it{
				{"-", token.Operator, 0, 1},
				{"==>", token.Operator, 2, 5},
				{"<>", token.Operator, 6, 8},
				{"<$>", token.Operator, 9, 12},
				{"%^", token.Operator, 13, 15},
				{"*&^%", token.Operator, 16, 20},
				{";|-", token.Operator, 21, 24},
				{"??", token.Operator, 25, 27},
			},
		},
	}
	matchAllTestWithTable(t, &table)
}

func TestErrorRecovery(t *testing.T) {
	source := "123abcd345 hello \"aaa\nagain"
	l := NewLexer(strings.NewReader(source), func(_ span.Position, _ string) {})
	expect := []it{
		{"123abcd345", token.Error, 0, 10},
		{"hello", token.Identifier, 11, 16},
		{"aaa", token.Error, 17, 21},
		{"\n", token.NewLine, 21, 22},
		{"again", token.Identifier, 22, 27},
	}
	for _, want := range expect {
		got := l.Next()
		span := span.NewSpan(span.Position{Line: 0, Column: 0, Offset: want.B}, span.Position{Line: 0, Column: 0, Offset: want.E})
		wanttok := token.Token{Typ: want.T, Val: want.N, Span: &span}
		equal := got.Typ != wanttok.Typ || got.Val != wanttok.Val
		equal = equal || got.Span.Beg.Offset != wanttok.Span.Beg.Offset
		equal = equal || got.Span.End.Offset != wanttok.Span.End.Offset
		if equal {
			t.Errorf("Wrong token - want: %#v got: %#v, lpos: %v", wanttok, got, l.position)
		}
	}
	eof := l.Next()
	if eof.Typ != token.Eof {
		t.Errorf("Expected EOF token, got: %v", eof)
	}
}

func matchAllTestWithTable(t *testing.T, table *tablet) {
	for _, test := range *table {
		t.Run(test.source, func(t *testing.T) {
			l := NewLexer(strings.NewReader(test.source), func(pos span.Position, msg string) {
				t.Fatalf("Error on pos %v with msg %v", pos, msg)
			})
			for _, want := range test.toks {
				got := l.Next()
				span := span.NewSpan(span.Position{Line: 0, Column: 0, Offset: want.B}, span.Position{Line: 0, Column: 0, Offset: want.E})
				wanttok := token.Token{Typ: want.T, Val: want.N, Span: &span}
				equal := got.Typ != wanttok.Typ || got.Val != wanttok.Val
				equal = equal || got.Span.Beg.Offset != wanttok.Span.Beg.Offset
				equal = equal || got.Span.End.Offset != wanttok.Span.End.Offset
				if equal {
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
// test operators
// test special so : and parenthesis and all
