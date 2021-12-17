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
	expect *data.Code
}

func TestEmittingConstValues(t *testing.T) {
	test := etest{
		{
			"10",
			codeFromBytes(1, []byte{
				isa.Constant, 0, isa.Pop,
			}),
		},
		{
			"true",
			codeFromBytes(1, []byte{
				isa.Constant, 0, isa.Pop,
			}),
		},
		{
			"false",
			codeFromBytes(1, []byte{
				isa.Constant, 0, isa.Pop,
			}),
		},
		{
			"a",
			codeFromBytes(1, []byte{
				isa.LoadDyn, 0, 0,
				isa.Pop,
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
				isa.Pop,
			}),
		},
		{
			"if true:\n" +
				"  1\n" +
				"  1\n" +
				"else:\n" +
				"  2\n" +
				"  2\n",
			codeFromBytes(5, []byte{
				isa.Constant, 0,
				isa.JumpIfFalse, 0, 11,
				isa.Constant, 1,
				isa.Pop,
				isa.Constant, 2,
				isa.Jump, 0, 8,
				isa.Constant, 3,
				isa.Pop,
				isa.Constant, 4,
				isa.Pop,
			}),
		},
	}
	matchResults(t, &test)
}

func TestEmittingMultipleStatements(t *testing.T) {
	test := etest{
		{
			"true\nfalse\n",
			codeFromBytes(2, []byte{
				isa.Constant, 0,
				isa.Pop,
				isa.Constant, 1,
				isa.Pop,
			}),
		},
	}
	matchResults(t, &test)
}

func codeFromBytes(cc int, bb []byte) *data.Code {
	c := data.NewCode()
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
			e := NewEmitter("dummy", NewInterner())
			got, errs := e.Compile(c)
			if len(errs) > 0 {
				t.Errorf("Unexpected compilation errors %v", errs)
			}
			if !bytes.Equal(got.Instrs, test.expect.Instrs) {
				t.Logf("Want:\n%s", isa.DisassembleCode(test.expect))
				t.Logf("\nGot:\n%s\n", isa.DisassembleCode(got))
				t.FailNow()
			}
		})
	}
}
