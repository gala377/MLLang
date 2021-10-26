package vm

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/gala377/MLLang/codegen"
	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
	"github.com/gala377/MLLang/syntax"
)

func echo(t *testing.T, v data.Value) data.NativeFunc {
	return data.NewNativeFunc("echo", 1, func(vs ...data.Value) data.Value {
		if len(vs) != 1 {
			t.Errorf("echo expects one value, got %v", vs)
		} else if !v.Equal(vs[0]) {
			t.Errorf("Values don't match: Want=%s, Got=%s", v, vs[0])
		}
		return data.NewInt(0)
	})
}

func TestSimpleVm(t *testing.T) {
	source := "val a = 1\necho a"
	echo := echo(t, data.NewInt(1))
	runTest(t, source, echo)
}

func runTest(t *testing.T, src string, echo data.NativeFunc) {
	p := syntax.NewParser(strings.NewReader(src))
	ast := p.Parse()
	if len(p.Errors()) > 0 {
		log.Printf("Parsing error:")
		for _, e := range p.Errors() {
			log.Print(e)
		}
		t.FailNow()
	}
	e := codegen.NewEmitter(codegen.NewInterner())
	c, errs := e.Compile(ast)
	if len(errs) > 0 {
		panic("Compilation errors")
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Test panicked, compiled code")
			fmt.Print(isa.DisassembleCode(c))
			panic(r)
		}
	}()
	name := e.Interner().Intern("echo")
	vm := vmWithEcho(name, echo)

	vm.Interpret(c)
}

func vmWithEcho(name data.InternedString, echo data.NativeFunc) *Vm {
	k := data.NewSymbol(name)
	env := map[data.Symbol]data.Value{
		k: &echo,
	}
	global := NewEnv()
	global.vals = env
	vm := VmWithEnv(global)
	return &vm
}
