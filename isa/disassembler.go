package isa

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/gala377/MLLang/data"
)

var instNames = [...]string{
	Return:      "Return",
	Constant:    "Constant",
	Call:        "Call",
	Jump:        "Jump",
	JumpBack:    "JumbpBack",
	JumpIfFalse: "JumpIfFalse",
	LoadDyn:     "LoadDyn",
	LoadLocal:   "LoadLocal",
	Pop:         "Pop",
	DefGlobal:   "DefGlobal",
	DefLocal:    "DefLocal",
	Lambda:      "Lambda",
	PushNone:    "PushNone",
	StoreLocal:  "StoreLocal",
	StoreDyn:    "StoreDyn",
	LoadDeref:   "LoadDeref",
	StoreDeref:  "StoreDeref",
	MakeCell:    "MakeCell",
	MakeList:    "MakeList",
	MakeTuple:   "MakeTuple",
}

const opCount = len(instNames)

var instArguments = [opCount]int{
	Return:      0,
	Constant:    1,
	Constant2:   2,
	Call:        1,
	Jump:        2,
	JumpBack:    2,
	JumpIfFalse: 2,
	LoadDyn:     2,
	LoadLocal:   2,
	Pop:         0,
	DefGlobal:   2,
	DefLocal:    2,
	Lambda:      2,
	PushNone:    0,
	StoreLocal:  2,
	StoreDyn:    2,
	StoreDeref:  2,
	LoadDeref:   2,
	MakeList:    2,
	MakeTuple:   2,
}

type additionalInfoFunc = func(*data.Code, []byte) string

var instSpecificInfos = [opCount]additionalInfoFunc{
	Constant:    writeConstant,
	Constant2:   writeConstantWide,
	Jump:        writeUint16,
	JumpBack:    writeUint16,
	JumpIfFalse: writeUint16,
	LoadDyn:     writeConstantWide,
	LoadLocal:   writeConstantWide,
	DefGlobal:   writeConstantWide,
	DefLocal:    writeConstantWide,
	Lambda:      writeConstantWide,
	StoreLocal:  writeConstantWide,
	StoreDyn:    writeConstantWide,
	StoreDeref:  writeConstantWide,
	LoadDeref:   writeConstantWide,
	MakeList:    writeUint16,
	MakeTuple:   writeUint16,
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

func writeConstantWide(code *data.Code, args []byte) string {
	i := binary.BigEndian.Uint16(args)
	return fmt.Sprintf("%16s", code.Consts[i])
}

func writeUint16(code *data.Code, args []byte) string {
	o := binary.BigEndian.Uint16(args)
	return fmt.Sprintf("%16d", o)
}
