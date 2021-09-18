package isa

import (
	"testing"

	"github.com/gala377/MLLang/data"
)

func TestDisassembling(t *testing.T) {
	want := "0000 RETURN\n0001 RETURN\n"
	c := Code{
		instrs: []byte{Return, Return},
	}
	got := DisassembleCode(c)
	if want != got {
		t.Errorf("Wrong disassembling.\nWant:\n%s\nGot:\n%s", want, got)
	}
}

func TestDisassemblingArgs(t *testing.T) {
	want := "0000 CONSTANT(0)             123\n0002 RETURN\n"
	c := Code{
		instrs: []byte{Constant, 0, Return},
		consts: []data.Value{data.NewInt(123)},
	}
	got := DisassembleCode(c)
	if want != got {
		t.Errorf("Wrong disassembling.\nWant:\n%s\nGot:\n%s", want, got)
	}
}
