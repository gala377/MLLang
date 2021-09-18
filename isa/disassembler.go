package isa

import (
	"fmt"
	"strconv"
	"strings"
)

var instNames = [...]string{
	Return:   "RETURN",
	Constant: "CONSTANT",
}

var instArguments = [...]uint{
	Return:   0,
	Constant: 1,
}

func PrintCode(code Code, name string) {
	fmt.Printf("== %s ==\n%s", name, DisassembleCode(code))
}

func DisassembleCode(code Code) string {
	var c strings.Builder
	for i := 0; i < len(code.instrs); {
		di, o := DisassembleInstr(code, i)
		i += o
		c.WriteString(di)
		c.WriteRune('\n')
	}
	return c.String()
}

func DisassembleInstr(code Code, offset int) (string, int) {
	var b strings.Builder
	op := code.instrs[offset]
	b.WriteString(fmt.Sprintf("%04d %s", offset, instNames[op]))
	args := instArguments[op]
	if args > 0 {
		aa := make([]string, 0, args)
		for i := 1; i <= int(args); i++ {
			a := code.instrs[offset+i]
			aa = append(aa, strconv.Itoa(int(a)))
		}
		b.WriteString(fmt.Sprintf("(%s)", strings.Join(aa, ",")))
	}
	return b.String(), int(1 + args)
}
