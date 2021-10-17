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
	"github.com/gala377/MLLang/syntax"
	"github.com/gala377/MLLang/vm"
)

var verboseFlag = flag.Bool("verbose", false, "when set shows logs from the virtual machine")

func main() {
	flag.Parse()
	fmt.Printf("%v\n", flag.Args())
	f := getFile()
	log.SetOutput(ioutil.Discard)
	vm.Debug = *verboseFlag
	evaluateBuffer(f)
}

func evaluateBuffer(buff []byte) {
	p := syntax.NewParser(bytes.NewReader(buff))
	ast := p.Parse()
	if len(p.Errors()) > 0 {
		log.Printf("Parsing error:")
		for _, e := range p.Errors() {
			log.Print(e)
		}
		os.Exit(1)
	}
	e := codegen.NewEmitter()
	c := e.Compile(ast)
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
