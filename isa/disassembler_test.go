package isa

import "testing"

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
	want := "0000 CONSTANT(2)\n0002 RETURN\n"
	c := Code{
		instrs: []byte{Constant, 2, Return},
	}
	got := DisassembleCode(c)
	if want != got {
		t.Errorf("Wrong disassembling.\nWant:\n%s\nGot:\n%s", want, got)
	}
}
