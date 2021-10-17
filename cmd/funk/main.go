package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gala377/MLLang/codegen"
	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/syntax"
	"github.com/gala377/MLLang/vm"
)

func main() {
	f := getFile()
	log.SetOutput(ioutil.Discard)
	vm.Debug = false
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
	filename := os.Args[1]
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(fmt.Errorf("%s file not found", filename))
		os.Exit(1)
	}
	return buffer
}
