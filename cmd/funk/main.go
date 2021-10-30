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
var showAst = flag.Bool("dump_ast", false, "just parse the file and print the ast to stdout")

func main() {
	flag.Parse()
	f := getFile()
	vm.Debug = *verboseFlag
	if !*verboseFlag {
		log.SetOutput(ioutil.Discard)
	}
	if *showAst {
		sr := bytes.NewReader(f)
		p := syntax.NewParser(sr)
		ast := p.Parse()
		fmt.Printf("%s", ast)
		return
	}
	evaluateBuffer(f)
}

func evaluateBuffer(buff []byte) {
	i := codegen.NewInterner()
	c, err := codegen.Compile(buff, i)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	if *showCode {
		printCode(c)
		os.Exit(0)
	}
	s := bytes.NewReader(buff)
	vm := vmWithStdEnv(s, i)
	vm.Interpret(c)
}

func printCode(c *data.Code) {
	fmt.Println(isa.DisassembleCode(c))
	for _, c := range c.Consts {
		switch v := c.(type) {
		case *data.Function:
			name := ""
			if v.Name.Inner() != nil {
				name = *v.Name.Inner()
			}
			fmt.Printf("\n=======Function %s========\n", name)
			printCode(v.Body)
		}
	}
}

func vmWithStdEnv(source *bytes.Reader, i *codegen.Interner) *vm.Vm {
	env := make(map[data.Symbol]data.Value)
	for _, fn := range stdEnv {
		s := i.Intern(fn.name)
		nf := data.NewNativeFunc(fn.name, fn.arity, fn.f)
		env[data.NewSymbol(s)] = &nf
	}
	globals := data.EnvFromMap(env)
	vm := vm.VmWithEnv(source, globals)
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
