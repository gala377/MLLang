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
		Path   string
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
	vm.AddToGlobals(m.Name, data.RecordFromMap(fields))
}

func (fs *funkSource) Inject(vm *vm.Vm) {
	inter := vm.Interner()
	c, err := codegen.Compile(fs.Path, fs.Source, inter)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	vm.AddSource(fs.Path, bytes.NewReader(fs.Source))
	_, err = vm.Interpret(c)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

var StdEnv = [...]EnvironmentEntry{
	&funcEntry{"add", 2, add},
	&funcEntry{"sub", 2, sub},
	&funcEntry{"mul", 2, mul},
	&funcEntry{"div", 2, div},
	&funcEntry{"neg", 1, neg},
	&funcEntry{"mod", 2, modulo},
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
	&funcEntry{"record?", 1, isRecord},
	&funcEntry{"or", 2, boolOr},
	&funcEntry{"and", 2, boolAnd},
	&funcEntry{"concat", 2, strConcat},
	&funcEntry{"loadFile", 1, loadFile},
	&stringsModule,
	&preludeModule,
	&ioModule,
	&seqModule,
	&convModule,
	&timeModule,
	&httpModule,
	&inspectModule,
	&recordsModule,
	&funkSource{"@errors", funkErrors},
	&funkSource{"@iter", funkIter},
	&funkSource{"@prelude", funkPrelude},
	&funkSource{"@conv", funkConv},
	&funkSource{"@seq", funkSeq},
	&funkSource{"@struct", funkStruct},
	&funkSource{"@records", funkRecords},
	&funkSource{"@io", funkIo},
	&funkSource{"@multimethods", funkMultimethod},
	&funkSource{"@cf", funkCf},
	&funkSource{"@funcs", funcFuncs},
}
