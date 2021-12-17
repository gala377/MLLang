package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"

	"github.com/gala377/MLLang/cmd/funk/std"
	"github.com/gala377/MLLang/codegen"
	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
	"github.com/gala377/MLLang/syntax"
	"github.com/gala377/MLLang/vm"
)

var verboseFlag = flag.Bool("verbose", false, "when set shows logs from the virtual machine")
var showCode = flag.Bool("dump_bytecode", false, "just compiles the file and prints it to the stdout")
var showAst = flag.Bool("dump_ast", false, "just parse the file and print the ast to stdout")
var panicOnError = flag.Bool("panic_on_error", false, "runtime error will cause panic in the interpreter")
var profile = flag.String("profile", "", "start profiling and write data to file specified as a value of this flag")

var filePath = ""

func main() {
	flag.Parse()
	parsePositionalArgs()
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
	evaluateBuffer(filePath, f)
}

func parsePositionalArgs() {
	if flag.NArg() == 0 {
		panic("Expected one positional argument which is a file name")
	}
	filePath = flag.Arg(0)
}

func evaluateBuffer(path string, buff []byte) {
	i := codegen.NewInterner()
	c, err := codegen.Compile(path, buff, i)
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
	defer func() {
		if r := recover(); r != nil {
			if msg, ok := r.(string); ok && msg == "runtime error" {
				if !*panicOnError {
					return
				}
			}
			panic(r)
		}
	}()
	if *profile != "" {
		f, err := os.Create(*profile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		f2, err := os.Create(*profile + "heapprofile")
		if err != nil {
			log.Fatal(err)
		}
		defer f2.Close()
		pprof.StartCPUProfile(f)
		vm.Interpret(c)
		pprof.WriteHeapProfile(f2)
		defer pprof.StopCPUProfile()
	} else {
		vm.Interpret(c)
	}
}

func printCode(c *data.Code) {
	fmt.Println(isa.DisassembleCode(c))
	for _, c := range c.Consts {
		switch v := c.(type) {
		case *data.Closure:
			name := ""
			if v.Name.Inner() != nil {
				name = *v.Name.Inner()
			}
			fmt.Printf("\n=======Function %s========\n", name)
			printCode(v.Body)
		}
	}
}

func vmWithStdEnv(source *bytes.Reader, interner *codegen.Interner) *vm.Vm {
	vm := vm.NewVm(filePath, source, interner)
	for _, e := range std.StdEnv {
		e.Inject(&vm)
	}
	return &vm
}

func getFile() []byte {
	filename := filePath
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(fmt.Errorf("'%s' file not found", filename))
		os.Exit(1)
	}
	return buffer
}
