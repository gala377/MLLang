package isa

import (
	"fmt"
	"strings"
)

var instNames = [...]string{
	Return: "RETURN",
}

var simpleInst = map[Op]struct{}{
	Return: {},
}

func PrintCode(code Code, name string) {
	fmt.Printf("== %s ==\n%s", name, DisassembleCode(code))
}

func DisassembleCode(code Code) string {
	var c strings.Builder
	for i := 0; i < len(code); {
		di, o := DisassembleInstr(code, i)
		i += o
		c.WriteString(di)
		c.WriteRune('\n')
	}
	return c.String()
}

func DisassembleInstr(code Code, offset int) (string, int) {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%04d ", offset))
	op := code[offset]
	if _, ok := simpleInst[op]; ok {
		b.WriteString(instNames[op])
		return b.String(), offset + 1
	}
	b.WriteString("UNKNOWN")
	return b.String(), offset + 1
}
