package vm

import (
	"fmt"
	"strings"

	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
)

var Debug = true

type Vm struct {
	code *isa.Code
	// global instruction pointer or per function instruction pointer?
	ip int
	// modules []*module ? map[symbol]*module?
	// thisModule *module
	stack    []data.Value
	stackTop int
}

func NewVm() Vm {
	return Vm{
		code:     nil,
		ip:       0,
		stack:    make([]data.Value, 0),
		stackTop: 0,
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
			v := vm.pop()
			fmt.Print(v.String())
			return v, nil
		case isa.Constant:
			arg := vm.readByte()
			v := vm.code.GetConstant(arg)
			vm.push(v)
		case isa.Call:
			arity := int(vm.readByte())
			args := make([]data.Value, 0, arity)
			for i := 0; i < arity; i++ {
				args = append(args, vm.pop())
			}
			callee := vm.pop()
			fn, ok := callee.(data.Callable)
			if !ok {
				// todo exception?
				panic("Trying to call something that is not callable")
			}
			switch {
			case arity == fn.Arity():
				vm.push(fn.Call(args...))
			case arity < fn.Arity():
				vm.push(data.NewPartialApp(fn, args...))
			case arity > fn.Arity():
				// todo exception?
				panic("Supplied more arguments than function takes")
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
