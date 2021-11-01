package vm

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

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
		globals  *data.Env
		locals   *data.Env
		source   *bytes.Reader
	}
)

func NewVm(source *bytes.Reader) Vm {
	globals := data.NewEnv()
	locals := data.NewEnv()
	return Vm{
		code:     nil,
		ip:       0,
		stack:    make([]data.Value, 0),
		stackTop: 0,
		globals:  &globals,
		locals:   &locals,
		source:   source,
	}
}

func VmWithEnv(source *bytes.Reader, env data.Env) Vm {
	vm := NewVm(source)
	vm.globals = &env
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
			v := vm.pop()
			env, ok := vm.pop().(*data.Env)
			if !ok {
				panic("ICE: on return popped value is not an env")
			}
			code, ok := vm.pop().(*data.Code)
			if !ok {
				panic("ICE: on return popped value is not a code")
			}
			ip, ok := vm.pop().(data.Int)
			if !ok {
				panic("ICE: on return popped value is not an ip")
			}
			vm.locals = env
			vm.code = code
			vm.ip = ip.Val
			vm.push(v)
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
		case isa.JumpIfFalse:
			off := vm.readShort()
			cond := vm.pop()
			if Debug {
				fmt.Printf("JumpIfFalse: jumping by %d", off)
			}
			ab, ok := cond.(*data.Bool)
			if !ok {
				vm.bail("conditiona has to be a boolean")
			}
			if !ab.Val {
				vm.ip += int(off) - 3
			}
			if Debug {
				i, _ := isa.DisassembleInstr(vm.code, vm.ip, -1)
				fmt.Printf("Instruction after jump %s", i)
			}
		case isa.Jump:
			off := vm.readShort()
			if Debug {
				fmt.Printf("Jump: jumping by %d", off)
			}
			vm.ip += int(off) - 3
			if Debug {
				i, _ := isa.DisassembleInstr(vm.code, vm.ip, -1)
				fmt.Printf("Instruction after jump %s", i)
			}
		case isa.DefGlobal:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			vm.globals.Insert(s, vm.pop())
		case isa.DefLocal:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			vm.locals.Insert(s, vm.pop())
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
			v, t := vm.applyFunc(fn, reverse(args))
			switch t.Kind {
			case data.Returned:
				vm.push(v)
			case data.Call:
				vm.push(data.Int{Val: vm.ip})
				vm.push(vm.code)
				vm.push(vm.locals)
				vm.ip = 0
				vm.code = t.Code
				vm.locals = t.Env
			case data.Error:
				vm.bail(v.String())
			}
		case isa.LoadDyn:
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
		case isa.LoadLocal:
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
		case isa.Lambda:
			arg := vm.readShort()
			l := vm.getFunctionAt(arg)
			lenv := data.NewEnv()
			for key, val := range vm.locals.Vals {
				lenv.Vals[key] = val
			}
			l = data.NewLambda(&lenv, l.Args, l.Body)
			vm.push(l)
		case isa.PushNone:
			vm.push(data.None)
		case isa.StoreDyn:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			vm.globals.Insert(s, vm.pop())
		case isa.StoreLocal:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			vm.locals.Insert(s, vm.pop())
		case isa.MakeCell:
			vm.push(data.NewCell(vm.pop()))
		case isa.LoadDeref:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			c := vm.locals.Lookup(s)
			if ac, ok := c.(*data.Cell); ok {
				vm.push(ac.Get())
			} else {
				vm.bail("ICE: LoadDeref used not on cell")
			}
		case isa.StoreDeref:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			c := vm.locals.Lookup(s)
			if ac, ok := c.(*data.Cell); ok {
				ac.Set(vm.pop())
			} else {
				vm.bail("ICE: StoreDeref used not on cell")
			}
		case isa.MakeList:
			size := int(vm.readShort())
			vals := make([]data.Value, 0, size)
			for i := 0; i < size; i++ {
				vals = append(vals, vm.pop())
			}
			l := data.NewList(reverse(vals))
			vm.push(l)
		case isa.MakeTuple:
			size := int(vm.readShort())
			vals := make([]data.Value, 0, size)
			for i := 0; i < size; i++ {
				vals = append(vals, vm.pop())
			}
			l := data.NewTuple(reverse(vals))
			vm.push(l)
		default:
			instr, _ := isa.DisassembleInstr(vm.code, vm.ip-1, -1)
			vm.bail(fmt.Sprintf("usupported command:\n%s", instr))
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
	s, _ := isa.DisassembleInstr(vm.code, vm.ip, vm.code.Lines[vm.ip])
	fmt.Println(s)
}

func (vm *Vm) applyFunc(fn data.Callable, args []data.Value) (data.Value, data.Trampoline) {
	argc := len(args)
	switch {
	case argc == fn.Arity():
		return fn.Call(args...)
	case argc < fn.Arity():
		if Debug {
			fmt.Printf("Partial application %d < %d", argc, fn.Arity())
		}
		return data.NewPartialApp(fn, args...), data.ReturnTramp
	case argc > fn.Arity():
		vm.bail("supplied more arguments than the function takes")
		return nil, data.Trampoline{}
	}
	panic("unreachable")
}

func (vm *Vm) getSymbolAt(i uint16) data.Symbol {
	s := vm.code.GetConstant2(i)
	if as, ok := s.(data.Symbol); ok {
		return as
	}
	vm.bail("expected constant to be a symbol")
	return data.Symbol{}
}

func (vm *Vm) getFunctionAt(i uint16) *data.Function {
	s := vm.code.GetConstant2(i)
	if as, ok := s.(*data.Function); ok {
		return as
	}
	vm.bail("expected constant to be a function")
	return &data.Function{}
}

func (vm *Vm) bail(msg string) {
	line := vm.code.Lines[vm.ip] + 1
	code := getLine(line, vm.source)

	fmt.Printf("\n\nRuntime error at line %d\n\n", line)
	fmt.Printf("%s\n\n", code)
	fmt.Println(msg)
	panic("\nruntime error\n")
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
