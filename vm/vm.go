package vm

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
)

var Debug = true

type (
	returnKind = byte

	trampoline struct {
		kind returnKind
	}

	Vm struct {
		code *isa.Code
		// global instruction pointer or per function instruction pointer?
		ip int
		// modules []*module ? map[symbol]*module?
		// thisModule *module
		stack    []data.Value
		stackTop int
		globals  Env
	}
)

const (
	Proceed returnKind = iota
	NormalCall
	TailCall
)

func NewVm() Vm {
	return Vm{
		code:     nil,
		ip:       0,
		stack:    make([]data.Value, 0),
		stackTop: 0,
		globals:  NewEnv(),
	}
}

func (vm *Vm) Interpret(code *isa.Code) (data.Value, error) {
	vm.code = code
	for {
		if vm.ip == vm.code.Len() {
			break
		}
		if Debug {
			fmt.Printf("Interpreter state for ip %d", vm.ip)
			vm.printInstr()
			vm.printStack()
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
		case isa.Call:
			arity := int(vm.readByte())
			args := make([]data.Value, 0, arity)
			for i := 0; i < arity; i++ {
				args = append(args, vm.pop())
			}
			callee := vm.pop()
			fn, ok := callee.(data.Callable)
			if !ok {
				panic("Trying to call something that is not callable")
			}
			vm.applyFunc(fn, reverse(args))
		case isa.DynLookup:
			// todo: add more lookups
			// now it only works for globals
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			if v := vm.globals.Lookup(s); v != nil {
				vm.push(v)
			} else {
				panic(fmt.Sprintf("variable %s undefined", s))
			}
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
}

func (vm *Vm) pop() data.Value {
	v := vm.stack[len(vm.stack)-1]
	vm.stackTop--
	assert(vm.stackTop > -1, "Stack top should never be less than 0")
	return v
}

func (vm *Vm) printStack() {
	v := make([]string, 0, vm.stackTop)
	for i := 0; i < vm.stackTop; i++ {
		v = append(v, vm.stack[i].String())
	}
	fmt.Printf("[%s]", strings.Join(v, ","))
}

func (vm *Vm) printInstr() {
	isa.DisassembleInstr(vm.code, vm.ip, vm.code.Lines[vm.ip])
}

func (vm *Vm) applyFunc(fn data.Callable, args []data.Value) trampoline {
	arity := fn.Arity()
	switch {
	case arity == fn.Arity():
		return vm.call(fn, args)
	case arity < fn.Arity():
		vm.push(data.NewPartialApp(fn, args...))
		return trampoline{kind: Proceed}
	case arity > fn.Arity():
		panic("supplied more arguments than the function takes")
	}
	panic("unreachable")
}

func (vm *Vm) call(fn data.Callable, args []data.Value) trampoline {
	switch c := fn.(type) {
	case *data.NativeFunc:
		vm.push(c.Call(args...))
		return trampoline{kind: Proceed}
	}
	panic("Unreachable")
}

func (vm *Vm) getSymbolAt(i uint16) data.Symbol {
	s := vm.code.GetConstant2(i)
	if as, ok := s.(data.Symbol); ok {
		return as
	}
	panic("Expected constant to be a symbol")
}

func reverse(s []data.Value) []data.Value {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
