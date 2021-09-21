package vm

import (
	"fmt"

	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/isa"
)

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
		}
	}
	return data.None, nil
}

func (vm *Vm) readByte() byte {
	b := vm.code.ReadByte(vm.ip)
	vm.ip += 1
	return b
}

func (vm *Vm) push(v data.Value) {
	vm.stack = append(vm.stack, v)
	vm.stackTop += 1
}

func (vm *Vm) pop() data.Value {
	v := vm.stack[len(vm.stack)-1]
	vm.stackTop -= 1
	assert(vm.stackTop > -1, "Stack top should never be less than 0")
	return v
}
