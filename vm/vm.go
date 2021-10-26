package vm

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/gala377/MLLang/code"
	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
)

var Debug = true

type (
	Vm struct {
		code *data.Code
		// global instruction pointer or per function instruction pointer?
		ip int
		// modules []*module ? map[symbol]*module?
		// thisModule *module
		stack    []data.Value
		stackTop int
		globals  data.Env
		locals   data.Env
		source   *bytes.Reader
	}
)

func NewVm(source *bytes.Reader) Vm {
	return Vm{
		code:     nil,
		ip:       0,
		stack:    make([]data.Value, 0),
		stackTop: 0,
		globals:  data.NewEnv(),
		locals:   data.NewEnv(),
		source:   source,
	}
}

func VmWithEnv(source *bytes.Reader, env data.Env) Vm {
	vm := NewVm(source)
	vm.globals = env
	return vm
}

func (vm *Vm) Interpret(code *data.Code) (data.Value, error) {
	vm.code = code
	for {
		if vm.ip == vm.code.Len() {
			break
		}
		if Debug {
			fmt.Println("======================+==========================")
			fmt.Printf("Interpreter state for ip %d\n", vm.ip)
			vm.printInstr()
			vm.printStack()
			fmt.Println("")
		}
		i := vm.readByte()
		switch i {
		case isa.Return:
			// restore ip, code and who knows what else
			v := vm.pop()
			fmt.Print(v.String())
			return v, nil
		case isa.Constant:
			arg := vm.readByte()
			v := vm.code.GetConstant(arg)
			vm.push(v)
		case isa.Constant2:
			arg := vm.readShort()
			v := vm.code.GetConstant2(arg)
			vm.push(v)
		case isa.Pop:
			vm.pop()
		case isa.DefGlobal:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			vm.globals.Instert(s, vm.pop())
		case isa.DefLocal:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			vm.locals.Instert(s, vm.pop())
		case isa.Call:
			arity := int(vm.readByte())
			args := make([]data.Value, 0, arity)
			for i := 0; i < arity; i++ {
				args = append(args, vm.pop())
			}
			callee := vm.pop()
			fn, ok := callee.(data.Callable)
			if !ok {
				vm.bail("Trying to call something that is not callable")
			}
			if Debug {
				fmt.Printf("Calling a function %s\n", fn.String())
			}
			vm.applyFunc(fn, reverse(args))
		case isa.DynLookup:
			// todo: add more lookups
			// now it only works for globals
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			if Debug {
				fmt.Printf("Global lookup of value %s\n", s)
			}
			v := vm.globals.Lookup(s)
			if v == nil {
				vm.bail(fmt.Sprintf("variable %s undefined", s))
			}
			if Debug {
				fmt.Printf("Lookup successful. Value is %s\n", v)
			}
			vm.push(v)
		case isa.LocalLookup:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			if Debug {
				fmt.Printf("Local lookup of value %s\n", s)
			}
			v := vm.locals.Lookup(s)
			if v == nil {
				vm.bail(fmt.Sprintf("variable %s undefined", s))
			}
			vm.push(v)
		}
	}
	return data.None, nil
}

func (vm *Vm) readByte() byte {
	b := vm.code.ReadByte(vm.ip)
	vm.ip++
	return b
}

func (vm *Vm) readShort() uint16 {
	args := []byte{vm.readByte(), vm.readByte()}
	index := binary.BigEndian.Uint16(args)
	return index
}

func (vm *Vm) push(v data.Value) {
	vm.stack = append(vm.stack, v)
	vm.stackTop++
	if Debug {
		fmt.Printf("Pushing value %s\nStack top is %d\n", v, vm.stackTop)
	}
}

func (vm *Vm) pop() data.Value {
	vm.stackTop--
	assert(vm.stackTop > -1, "Stack top should never be less than 0")
	v := vm.stack[vm.stackTop]
	vm.stack = vm.stack[:vm.stackTop]
	if Debug {
		fmt.Printf("Popping value %s\nStack top is %d\n", v, vm.stackTop)
	}
	return v
}

func (vm *Vm) printStack() {
	v := make([]string, 0, vm.stackTop)
	for i := 0; i < vm.stackTop; i++ {
		v = append(v, vm.stack[i].String())
	}
	fmt.Printf("[%s]\n", strings.Join(v, ", "))
}

func (vm *Vm) printInstr() {
	s, _ := code.DisassembleInstr(vm.code, vm.ip, vm.code.Lines[vm.ip])
	fmt.Println(s)
}

func (vm *Vm) applyFunc(fn data.Callable, args []data.Value) data.Trampoline {
	argc := len(args)
	switch {
	case argc == fn.Arity():
		return vm.call(fn, args)
	case argc < fn.Arity():
		if Debug {
			fmt.Printf("Parital application %d < %d", argc, fn.Arity())
		}
		vm.push(data.NewPartialApp(fn, args...))
		return data.ProceedTramp
	case argc > fn.Arity():
		vm.bail("supplied more arguments than the function takes")
		return data.Trampoline{}
	}
	panic("unreachable")
}

func (vm *Vm) call(fn data.Callable, args []data.Value) data.Trampoline {
	switch c := fn.(type) {
	case *data.NativeFunc, *data.PartialApp:
		v, t := c.Call(args...)
		vm.push(v)
		return t
	}
	panic("Unreachable")
}

func (vm *Vm) getSymbolAt(i uint16) data.Symbol {
	s := vm.code.GetConstant2(i)
	if as, ok := s.(data.Symbol); ok {
		return as
	}
	vm.bail("expected constant to be a symbol")
	return data.Symbol{}
}

func (vm *Vm) bail(msg string) {
	line := vm.code.Lines[vm.ip+1]
	code := getLine(line, vm.source)

	fmt.Printf("\n\nRuntime error at line %d\n\n", line)
	fmt.Printf("%s\n\n", code)
	fmt.Println(msg)
	panic("runtime error")
}

func reverse(s []data.Value) []data.Value {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

type seekReader interface {
	io.Reader
	io.Seeker
}

func getLine(line int, r seekReader) string {
	r.Seek(0, 0)
	sc := bufio.NewScanner(r)
	sc.Split(bufio.ScanLines)
	for cline := 1; sc.Scan(); cline++ {
		if cline == line {
			return sc.Text()
		}
	}
	panic("could not find given line")
}
