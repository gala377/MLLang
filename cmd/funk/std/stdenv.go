package std

import (
	"bytes"
	"fmt"
	"os"

	"github.com/gala377/MLLang/codegen"
	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/vm"
)

type (
	EnvironmentEntry interface {
		Inject(vm *vm.Vm)
	}

	AsValue interface {
		AsValue(vm *vm.Vm) data.Value
	}

	funcEntry struct {
		Name  string
		Arity int
		F     func(data.VmProxy, ...data.Value) (data.Value, error)
	}

	module struct {
		Name    string
		Entries map[string]AsValue
	}

	funkSource struct {
		Source []byte
	}
)

func (f *funcEntry) Inject(vm *vm.Vm) {
	vm.AddToGlobals(f.Name, f.AsValue(vm))
}

func (f *funcEntry) AsValue(vm *vm.Vm) data.Value {
	return data.NewNativeFunc(f.Name, f.Arity, f.F)
}

func (m *module) Inject(vm *vm.Vm) {
	fields := map[data.Symbol]data.Value{}
	for key, val := range m.Entries {
		fields[vm.CreateSymbol(key)] = val.AsValue(vm)
	}
	vm.AddToGlobals(m.Name, data.NewRecord(fields))
}

func (fs *funkSource) Inject(vm *vm.Vm) {
	inter := vm.Interner()
	c, err := codegen.Compile(fs.Source, inter)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	olds := vm.ReplaceSource(bytes.NewReader(fs.Source))
	_, err = vm.Interpret(c)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	vm.ReplaceSource(olds)
}

var StdEnv = [...]EnvironmentEntry{
	&funcEntry{"add", 2, add},
	&funcEntry{"lt?", 2, lessThan},
	&funcEntry{"eq?", 2, equal},
	&funcEntry{"not", 1, not},
	&funcEntry{"panic", 1, vmPanic},
	&funcEntry{"gensym", 0, vmGenSym},
	&funcEntry{"spawn", 1, vmSpawn},
	&funcEntry{"int?", 1, isInt},
	&funcEntry{"list?", 1, isList},
	&funcEntry{"tuple?", 1, isTuple},
	&funcEntry{"seq?", 1, isSeq},
	&funcEntry{"string?", 1, isString},
	&funcEntry{"symbol?", 1, isSymbol},
	&funcEntry{"function?", 1, isFunction},
	&funcEntry{"float?", 1, isFloat},
	&funcEntry{"bool?", 1, isBool},
	&module{
		Name: "io",
		Entries: map[string]AsValue{
			"print":  &funcEntry{"print", 1, print},
			"printf": &funcEntry{"printf", 2, printf},
		},
	},
	&seqModule,
	&convModule,
	&timeModule,
	&httpModule,
	&funkSource{Source: funkPrelude},
	&funkSource{Source: funkConv},
	&funkSource{Source: funkSeq},
	&funkSource{Source: funkStruct},
}
