package codegen

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
	"github.com/gala377/MLLang/syntax"
)

type etest []struct {
	source string
	expect *isa.Code
}

func TestEmittingConstValues(t *testing.T) {
	test := etest{
		{
			"10",
			codeFromBytes(1, []byte{
				isa.Constant, 0, isa.Return,
			}),
		},
		{
			"true",
			codeFromBytes(1, []byte{
				isa.Constant, 0, isa.Return,
			}),
		},
		{
			"false",
			codeFromBytes(1, []byte{
				isa.Constant, 0, isa.Return,
			}),
		},
		{
			"a",
			codeFromBytes(1, []byte{
				isa.DynLookup, 0, 0,
				isa.Return,
			}),
		},
	}
	matchResults(t, &test)
}

func TestEmittingIf(t *testing.T) {
	test := etest{
		{
			"if true:\n" +
				"  1\n" +
				"else:\n" +
				"  2",
			codeFromBytes(3, []byte{
				isa.Constant, 0,
				isa.JumpIfFalse, 0, 8,
				isa.Constant, 1,
				isa.Jump, 0, 5,
				isa.Constant, 2,
				isa.Return,
			}),
		},
	}
	matchResults(t, &test)
}

func codeFromBytes(cc int, bb []byte) *isa.Code {
	c := isa.NewCode()
	c.Instrs = bb
	consts := []data.Value{}
	for i := 0; i < cc; i++ {
		consts = append(consts, data.None)
	}
	c.Consts = consts
	c.Lines = make([]int, len(bb))
	return &c
}

func matchResults(t *testing.T, table *etest) {
	for _, test := range *table {
		t.Run(test.source, func(t *testing.T) {
			t.Logf("SOURCE IS %v", test.source)
			p := syntax.NewParser(strings.NewReader(test.source))
			c := p.Parse()
			e := NewEmitter()
			got := e.Compile(c)
			if !bytes.Equal(got.Instrs, test.expect.Instrs) {
				t.Logf("Want:\n%s", isa.DisassembleCode(test.expect))
				t.Logf("\nGot:\n%s\n", isa.DisassembleCode(got))
				t.FailNow()
			}
		})
	}
}
