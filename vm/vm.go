package vm

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/gala377/MLLang/codegen"
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
		interner *codegen.Interner
		gensymc  uint
	}
)

func NewVm(source *bytes.Reader, interner *codegen.Interner) Vm {
	globals := data.NewEnv()
	locals := data.NewEnv()
	return Vm{
		code:     nil,
		ip:       0,
		stack:    make([]data.Value, 0),
		stackTop: 0,
		globals:  globals,
		locals:   locals,
		source:   source,
		interner: interner,
		gensymc:  0,
	}
}

func VmWithEnv(source *bytes.Reader, interner *codegen.Interner, env *data.Env) Vm {
	vm := NewVm(source, interner)
	vm.globals = env
	return vm
}

func (vm *Vm) AddToGlobals(name string, v data.Value) {
	vm.globals.Insert(vm.CreateSymbol(name), v)
}

func (vm *Vm) CreateSymbol(s string) data.Symbol {
	is := vm.interner.Intern(s)
	return data.NewSymbol(is)
}

func (vm *Vm) Interner() *codegen.Interner {
	return vm.interner
}

// todo: remove later. It's here for now so
// that we can easily inject std env fnk source files
// however when we introduce importing
// it should not be neccessary.
// Maybe code object should carry pointer to source?
func (vm *Vm) ReplaceSource(ns *bytes.Reader) *bytes.Reader {
	old := vm.source
	vm.source = ns
	return old
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
			if vm.stackTop == 0 {
				// top level return
				return v, nil
			}
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
			ab, ok := cond.(data.Bool)
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
		case isa.JumpBack:
			off := vm.readShort()
			if Debug {
				fmt.Printf("JumpBack: jumping by %d", off)
			}
			vm.ip -= int(off) + 3
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
		case isa.Call0:
			callee, ok := vm.pop().(data.Callable)
			if !ok {
				vm.bail(fmt.Sprintf("Cannot call %s", callee))
			}
			if callee.Arity() != 0 {
				vm.bail("Expected nullary callable")
			}
			vm.handleCall(callee.Call(vm))
		case isa.Call1:
			arg := vm.pop()
			callee, ok := vm.pop().(data.Callable)
			if !ok {
				vm.bail(fmt.Sprintf("Cannot apply %v", callee))
			}
			if callee.Arity() == 0 {
				vm.bail("Expected non nullary callable")
			}
			vm.handleCall(vm.apply1(callee, arg))
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
		case isa.Closure:
			arg := vm.readShort()
			l := vm.getFunctionAt(arg)
			lenv := data.NewEnv()
			for key, val := range vm.locals.Vals {
				lenv.Vals[key] = val
			}
			l = data.NewLambda(l.Name, lenv, l.Args, l.Body)
			vm.push(l)
		case isa.PushNone:
			vm.push(data.None)
		case isa.StoreDyn:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			err := vm.globals.Set(s, vm.pop())
			if err != nil {
				vm.bail(err.Error())
			}
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
		case isa.MakeRecord:
			type Pair struct {
				k data.Symbol
				v data.Value
			}
			size := int(vm.readShort())
			pairs := make([]Pair, 0, size)
			for i := 0; i < size; i++ {
				key, ok := vm.pop().(data.Symbol)
				if !ok {
					vm.bail("record fields have to be symbols")
				}
				pairs = append(pairs, Pair{key, vm.pop()})
			}
			rec := data.EmptyRecord()
			for i := size - 1; i >= 0; i-- {
				rec.SetField(pairs[i].k, pairs[i].v)
			}
			vm.push(rec)
		case isa.GetField:
			name := vm.getSymbolAt(vm.readShort())
			rec, ok := vm.pop().(*data.Record)
			if !ok {
				vm.bail("field can only be accessed on a record")
			}
			val, ok := rec.GetField(name)
			if !ok {
				vm.bail(fmt.Sprintf("missing field %s", name))
			}
			vm.push(val)
		case isa.SetField:
			val := vm.pop()
			name := vm.getSymbolAt(vm.readShort())
			rec, ok := vm.pop().(*data.Record)
			if !ok {
				vm.bail("field can only be accessed on a record")
			}
			rec.SetField(name, val)
		case isa.InstallHandler:
			argc := vm.readShort()
			arms := map[*data.Type]data.Callable{}
			for i := 0; i < int(argc); i++ {
				hfunc, ok := vm.pop().(data.Callable)
				if !ok {
					vm.bail("ICE: with clause handler is not a callable")
				}
				typ, ok := vm.pop().(*data.Type)
				if !ok {
					vm.bail("Handler can only switch on effect types")
				}
				arms[typ] = hfunc
			}
			handler := data.NewHandler(arms)
			vm.push(handler)
		case isa.PopHandler:
			ret := vm.pop()
			_, ok := vm.pop().(*data.Handler)
			if !ok {
				vm.bail("ICE PopHandler did not pop a handler")
			}
			vm.push(ret)
		default:
			instr, _ := isa.DisassembleInstr(vm.code, vm.ip-1, -1)
			vm.bail(fmt.Sprintf("usupported command:\n%s", instr))
		}
	}
	vm.ip = 0
	vm.code = nil
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
		return fn.Call(vm, args...)
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

// Specialised func application for applying only 1 argument.
// Allows to skip allocation of slice for arguments.
func (vm *Vm) apply1(fn data.Callable, arg data.Value) (data.Value, data.Trampoline) {
	switch {
	case fn.Arity() == 1:
		return fn.Call(vm, arg)
	case fn.Arity() > 1:
		if Debug {
			fmt.Printf("Partial application 1 < %d", fn.Arity())
		}
		return data.PartialApp1(fn, arg), data.ReturnTramp
	case fn.Arity() < 1:
		vm.bail("supplied more arguments than the function takes")
		return nil, data.Trampoline{}
	}
	panic("unreachable")
}

func (vm *Vm) handleCall(retval data.Value, tramp data.Trampoline) {
	for {
		switch tramp.Kind {
		case data.Returned:
			vm.push(retval)
		case data.Call:
			vm.push(data.Int{Val: vm.ip})
			vm.push(vm.code)
			vm.push(vm.locals)
			vm.ip = 0
			vm.code = tramp.Code
			vm.locals = tramp.Env
		case data.Error:
			vm.bail(retval.String())
		case data.Effect:
			args, ok := retval.(data.Tuple)
			if !ok {
				vm.bail("ICE: expected tuple of (type, arg) when performing an effect. Not a tuple")
			}
			typ, ok := vm.unsafeTupleGet(args, 0).(*data.Type)
			if !ok {
				vm.bail("ICE: expected tuple of (type, arg) when performing an effect. Not a type")
			}
			arg := vm.unsafeTupleGet(args, 1)
			retval, tramp = vm.handleEffect(typ, arg)
			continue
		}
		break
	}
}

func (vm *Vm) handleEffect(typ *data.Type, arg data.Value) (data.Value, data.Trampoline) {
	// captured reversed stack
	rstack := []data.Value{}
	// either bails or returns new function to call
	for {
		if vm.stackTop == 0 {
			vm.bail(fmt.Sprintf("Unhandled effect %s with val %s", typ, arg))
		}
		sv := vm.pop()
		handler, ok := sv.(*data.Handler)
		if ok {
			for ty, h := range handler.Clauses {
				if typ.Equal(ty) {
					switch h.Arity() {
					case 1:
						return h.Call(vm, arg)
					case 2:
						// todo
						vm.bail("Passing continuations to effects unsupported")
						fmt.Printf("Stack %v", rstack)

					default:
						vm.bail("ICE: unsupported arity of effect handling clause")
					}
				}
			}
		}
		rstack = append(rstack, sv)
	}
}

func (vm *Vm) getSymbolAt(i uint16) data.Symbol {
	s := vm.code.GetConstant2(i)
	if as, ok := s.(data.Symbol); ok {
		return as
	}
	vm.bail("expected constant to be a symbol")
	return data.Symbol{}
}

func (vm *Vm) getFunctionAt(i uint16) *data.Closure {
	s := vm.code.GetConstant2(i)
	if as, ok := s.(*data.Closure); ok {
		return as
	}
	vm.bail("expected constant to be a function")
	return &data.Closure{}
}

func (vm *Vm) bail(msg string) {
	ip := vm.ip
	if ip >= len(vm.code.Lines) {
		ip = len(vm.code.Lines) - 1
	}
	line := vm.code.Lines[ip] + 1
	code := getLine(line, vm.source)

	fmt.Printf("\n\nRuntime error at line %d\n\n", line)
	fmt.Printf("%s\n\n", code)
	fmt.Println(msg)
	panic("\nruntime error\n")
}

func (vm *Vm) Panic(msg string) {
	vm.bail(msg)
}

func (vm *Vm) GenerateSymbol() data.Symbol {
	vm.gensymc++
	str := fmt.Sprintf("@gensym[%d]", vm.gensymc)
	return data.NewSymbol(&str)
}

func (vm *Vm) Clone() data.VmProxy {
	loc := data.NewEnv()
	return &Vm{
		code:     nil,
		ip:       0,
		stack:    make([]data.Value, 0),
		stackTop: 0,
		globals:  vm.globals,
		locals:   loc,
		source:   vm.source,
		interner: vm.interner.Clone(),
		gensymc:  vm.gensymc,
	}
}

func (vm *Vm) RunClosure(c data.Callable, args ...data.Value) data.Value {
	v, t := c.Call(vm, args...)
	switch t.Kind {
	case data.Returned:
		return v
	case data.Error:
		panic(fmt.Sprintf("closure returned a top lovel error %s", v))
	case data.Call:
		c := t.Code
		v, err := vm.Interpret(c)
		if err != nil {
			panic(fmt.Sprintf("error in RunClosure: %s", err))
		}
		return v
	}
	return data.None
}

func (vm *Vm) unsafeTupleGet(t data.Tuple, i int) data.Value {
	v, err := t.Get(data.NewInt(i))
	if err != nil {
		vm.bail("ICE: tried to take a value from tuple past its indices")
	}
	return v
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
	return "could not find given line"
}
