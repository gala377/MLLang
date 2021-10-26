package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gala377/MLLang/codegen"
	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
	"github.com/gala377/MLLang/syntax"
	"github.com/gala377/MLLang/vm"
)

var verboseFlag = flag.Bool("verbose", false, "when set shows logs from the virtual machine")
var showCode = flag.Bool("dump_bytecode", false, "just compiles the file and prints it to the stdout")

func main() {
	flag.Parse()
	f := getFile()
	log.SetOutput(ioutil.Discard)
	vm.Debug = *verboseFlag
	evaluateBuffer(f)
}

func evaluateBuffer(buff []byte) {
	p := syntax.NewParser(bytes.NewReader(buff))
	ast := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Print("Parsing error:")
		for _, e := range p.Errors() {
			fmt.Print(e)
		}
		os.Exit(1)
	}
	e := codegen.NewEmitter()
	c, errs := e.Compile(ast)
	if len(errs) > 0 {
		panic("Compilation erros")
	}
	if *showCode {
		fmt.Println(isa.DisassembleCode(c))
		os.Exit(0)
	}
	vm := vmWithStdEnv(e.Interner())
	vm.Interpret(c)
}

func vmWithStdEnv(i *codegen.Interner) *vm.Vm {
	env := make(map[data.Symbol]data.Value)
	for _, fn := range stdEnv {
		s := i.Intern(fn.name)
		nf := data.NewNativeFunc(fn.name, fn.arity, fn.f)
		env[data.NewSymbol(s)] = &nf
	}
	globals := vm.EnvFromMap(env)
	vm := vm.VmWithEnv(globals)
	return &vm
}

func getFile() []byte {
	if flag.NArg() == 0 {
		panic("Expected one positional argument which is a file name")
	}
	filename := flag.Arg(0)
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(fmt.Errorf("'%s' file not found", filename))
		os.Exit(1)
	}
	return buffer
}
