package isa

import (
	"fmt"
	"strconv"
	"strings"
)

var instNames = [...]string{
	Return:   "RETURN",
	Constant: "CONSTANT",
	Call:     "Call",
	Jump:     "Jump",
	JumpIf:   "JumpIf",
}

const opCount = len(instNames)

var instArguments = [opCount]int{
	Return:   0,
	Constant: 1,
	Call:     1,
	Jump:     1,
	JumpIf:   1,
}

type additionalInfoFunc = func(*Code, []byte) string

var instSpecificInfos = [opCount]additionalInfoFunc{
	Constant: writeConstant,
}

func PrintCode(code *Code, name string) {
	fmt.Printf("== %s ==\n%s", name, DisassembleCode(code))
}

func DisassembleCode(code *Code) string {
	var c strings.Builder
	line := -1
	for i := 0; i < len(code.Instrs); {
		di, o := DisassembleInstr(code, i, line)
		line = code.Lines[i]
		i += o
		c.WriteString(di)
		c.WriteRune('\n')
	}
	return c.String()
}

func DisassembleInstr(code *Code, offset int, lline int) (string, int) {
	var b strings.Builder
	op := code.Instrs[offset]
	line := "    |"
	if lline != code.Lines[offset] {
		line = fmt.Sprintf("%5d", code.Lines[offset])
	}
	b.WriteString(fmt.Sprintf("%04d %s %s", offset, line, instNames[op]))
	args := instArguments[op]
	if args > 0 {
		aa := make([]string, 0, args)
		for i := 1; i <= int(args); i++ {
			a := code.Instrs[offset+i]
			aa = append(aa, strconv.Itoa(int(a)))
		}
		b.WriteString(fmt.Sprintf("(%s)", strings.Join(aa, ",")))
		if s := instSpecificInfos[op]; s != nil {
			b.WriteString(s(code, code.Instrs[offset+1:offset+args+1]))
		}
	}
	return b.String(), 1 + args
}

func writeConstant(code *Code, args []byte) string {
	return fmt.Sprintf("%16s", code.Consts[args[0]])
}
