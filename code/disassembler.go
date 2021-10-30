package code

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
)

var instNames = [...]string{
	isa.Return:      "Return",
	isa.Constant:    "Constant",
	isa.Call:        "Call",
	isa.Jump:        "Jump",
	isa.JumpIfFalse: "JumpIfFalse",
	isa.LoadDyn:     "LoadDyn",
	isa.LoadLocal:   "LoadLocal",
	isa.Pop:         "Pop",
	isa.DefGlobal:   "DefGlobal",
	isa.DefLocal:    "DefLocal",
	isa.Lambda:      "Lambda",
	isa.PushNone:    "PushNone",
	isa.StoreLocal:  "StoreLocal",
	isa.StoreDyn:    "StoreDyn",
}

const opCount = len(instNames)

var instArguments = [opCount]int{
	isa.Return:      0,
	isa.Constant:    1,
	isa.Constant2:   2,
	isa.Call:        1,
	isa.Jump:        2,
	isa.JumpIfFalse: 2,
	isa.LoadDyn:     2,
	isa.LoadLocal:   2,
	isa.Pop:         0,
	isa.DefGlobal:   2,
	isa.DefLocal:    2,
	isa.Lambda:      2,
	isa.PushNone:    0,
	isa.StoreLocal:  2,
	isa.StoreDyn:    2,
}

type additionalInfoFunc = func(*data.Code, []byte) string

var instSpecificInfos = [opCount]additionalInfoFunc{
	isa.Constant:    writeConstant,
	isa.Constant2:   writeConstant2,
	isa.Jump:        writeJump,
	isa.JumpIfFalse: writeJump,
	isa.LoadDyn:     writeConstant2,
	isa.LoadLocal:   writeConstant2,
	isa.DefGlobal:   writeConstant2,
	isa.DefLocal:    writeConstant2,
	isa.Lambda:      writeConstant2,
	isa.StoreLocal:  writeConstant2,
	isa.StoreDyn:    writeConstant2,
}

func PrintCode(code *data.Code, name string) {
	fmt.Printf("== %s ==\n%s", name, DisassembleCode(code))
}

func DisassembleCode(code *data.Code) string {
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

func DisassembleInstr(code *data.Code, offset int, lline int) (string, int) {
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

func writeConstant(code *data.Code, args []byte) string {
	return fmt.Sprintf("%16s", code.Consts[args[0]])
}

func writeConstant2(code *data.Code, args []byte) string {
	i := binary.BigEndian.Uint16(args)
	return fmt.Sprintf("%16s", code.Consts[i])
}

func writeJump(code *data.Code, args []byte) string {
	o := binary.BigEndian.Uint16(args)
	return fmt.Sprintf("%16d", o)
}
