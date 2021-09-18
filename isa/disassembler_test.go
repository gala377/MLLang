package isa

import "testing"

func TestDisassembling(t *testing.T) {
	want := "0000 RETURN\n0001 RETURN\n"
	c := Code{Return, Return}
	got := DisassembleCode(c)
	if want != got {
		t.Errorf("Wring disassembling.\nWant:\n%s\nGot:\n%s", want, got)
	}
}
