package isa

import (
	"testing"

	"github.com/gala377/MLLang/data"
)

func TestDisassembling(t *testing.T) {
	want := "0000     1 RETURN\n0001     2 RETURN\n"
	c := &Code{
		instrs: []byte{Return, Return},
		lines:  []int{1, 2},
	}
	got := DisassembleCode(c)
	if want != got {
		t.Errorf("Wrong disassembling.\nWant:\n%s\nGot:\n%s", want, got)
	}
}

func TestDisassemblingArgs(t *testing.T) {
	want := "0000     1 CONSTANT(0)             123\n0002     | RETURN\n"
	c := &Code{
		instrs: []byte{Constant, 0, Return},
		consts: []data.Value{data.NewInt(123)},
		lines:  []int{1, 1, 1},
	}
	got := DisassembleCode(c)
	if want != got {
		t.Errorf("Wrong disassembling.\nWant:\n%s\nGot:\n%s", want, got)
	}
}
