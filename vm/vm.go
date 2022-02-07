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

const FUNC_FRAME_SIZE = 3

type (
	Vm struct {
		code  *data.Code
		ip    int
		stack []data.Value
		// shows how many elements are currently in the stack.
		// stackTop == 0 means an empty stack
		stackTop int
		globals  *data.Env
		locals   *data.Env
		interner *codegen.Interner
		gensymc  uint

		// for better error messages
		sources map[string]*bytes.Reader
	}
)

func NewVm(path string, source *bytes.Reader, interner *codegen.Interner) Vm {
	globals := data.NewEnv()
	locals := data.NewEnv()
	sources := map[string]*bytes.Reader{
		path: source,
	}
	return Vm{
		code:     nil,
		ip:       0,
		stack:    make([]data.Value, 0, 1024),
		stackTop: 0,
		globals:  globals,
		locals:   locals,
		interner: interner,
		gensymc:  0,
		sources:  sources,
	}
}

func VmWithEnv(path string, source *bytes.Reader, interner *codegen.Interner, env *data.Env) Vm {
	vm := NewVm(path, source, interner)
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

func (vm *Vm) AddSource(path string, s *bytes.Reader) {
	vm.sources[path] = s
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
			vm.ip, vm.code, vm.locals = vm.popFunctionFrame()
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
		case isa.Rotate:
			top := vm.stackTop - 1
			vm.stack[top], vm.stack[top-1] = vm.stack[top-1], vm.stack[top]
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
		case isa.TailCall0:
			callee, ok := vm.pop().(data.Callable)
			if !ok {
				vm.bail(fmt.Sprintf("Cannot call %s", callee))
			}
			if callee.Arity() != 0 {
				vm.bail("Expected nullary callable")
			}
			v, t := callee.Call(vm)
			vm.handleCall(v, t, true)
		case isa.Call0:
			callee, ok := vm.pop().(data.Callable)
			if !ok {
				vm.bail(fmt.Sprintf("Cannot call %s", callee))
			}
			if callee.Arity() != 0 {
				vm.bail("Expected nullary callable")
			}
			v, t := callee.Call(vm)
			vm.handleCall(v, t, false)
		case isa.TailCall1:
			arg := vm.pop()
			callee, ok := vm.pop().(data.Callable)
			if !ok {
				vm.bail(fmt.Sprintf("Cannot apply %v", callee))
			}
			if callee.Arity() == 0 {
				vm.bail("Expected non nullary callable")
			}
			v, t := vm.apply1(callee, arg)
			vm.handleCall(v, t, true)
		case isa.Call1:
			arg := vm.pop()
			callee, ok := vm.pop().(data.Callable)
			if !ok {
				vm.bail(fmt.Sprintf("Cannot apply %v", callee))
			}
			if callee.Arity() == 0 {
				vm.bail("Expected non nullary callable")
			}
			v, t := vm.apply1(callee, arg)
			vm.handleCall(v, t, false)
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
				vm.bail("IEE: LoadDeref used not on cell")
			}
		case isa.StoreDeref:
			arg := vm.readShort()
			s := vm.getSymbolAt(arg)
			c := vm.locals.Lookup(s)
			if ac, ok := c.(*data.Cell); ok {
				ac.Set(vm.pop())
			} else {
				vm.bail("IEE: StoreDeref used not on cell")
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
					vm.bail("IEE: with clause handler is not a callable")
				}
				typ, ok := vm.pop().(*data.Type)
				if !ok {
					vm.bail("Handler can only switch on effect types")
				}
				if _, ok := arms[typ]; ok {
					vm.bail("Handler already has an arm for the type %s", typ.Name.String())
				}
				arms[typ] = hfunc
			}
			handler := data.NewHandler(arms)
			vm.push(handler)
		case isa.PopHandler:
			ret := vm.pop()
			_, ok := vm.pop().(*data.Handler)
			if !ok {
				vm.bail("IEE PopHandler did not pop a handler")
			}
			vm.push(ret)
		case isa.MakeEffect:
			name, ok := vm.pop().(data.Symbol)
			if !ok {
				vm.bail("IEE Expected symbol on the stack to make an effect")
			}
			vm.push(data.NewType(name))
		case isa.TailResume0, isa.TailResume1:
			var arg data.Value = data.None
			if i == isa.TailResume1 {
				arg = vm.pop()
			}
			cont, ok := vm.pop().(*data.Continuation)
			if !ok {
				vm.bail("resume expression expects a continuation to call")
			}
			ip, code, env := vm.popFunctionFrame()
			vm.push(cont.Handler)
			if code.Instrs[ip-1] != isa.PopHandler {
				vm.bail("Cannot tail resume. Tail resumption only works if it's" +
					" the last expression in the with clause or the function that" +
					" tail resumes is in tail position for the with clause")
			}
			vm.ip = ip - 1
			vm.code = code
			vm.locals = env
			v, t := cont.Call(vm, arg)
			vm.handleCall(v, t, false)
		case isa.Resume:
			cont, ok := vm.pop().(*data.Continuation)
			if !ok {
				vm.bail("resume expression expects a continuation to call")
			}
			vm.push(cont.Handler)
			vm.push(cont)
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

func (vm *Vm) handleCall(retval data.Value, tramp data.Trampoline, tailcall bool) {
	for {
		switch tramp.Kind {
		case data.Returned:
			vm.push(retval)
		case data.Call:
			if !tailcall {
				vm.push(data.Int{Val: vm.ip})
				vm.push(vm.code)
				vm.push(vm.locals)
			}
			vm.ip = 0
			vm.code = tramp.Code
			vm.locals = tramp.Env
		case data.Error:
			vm.bail(retval.String())
		case data.Effect:
			args, ok := retval.(data.Tuple)
			if !ok {
				vm.bail("IEE: expected tuple of (type, arg) when performing an effect. Not a tuple")
			}
			typ, ok := vm.unsafeTupleGet(args, 0).(*data.Type)
			if !ok {
				vm.bail("IEE: expected tuple of (type, arg) when performing an effect. Not a type")
			}
			arg := vm.unsafeTupleGet(args, 1)
			retval, tramp, tailcall = vm.handleEffect(typ, arg)
			continue
		case data.RestoreContinuation:
			// push current frame
			vm.push(data.Int{Val: vm.ip})
			vm.push(vm.code)
			vm.push(vm.locals)
			args, ok := retval.(data.Tuple)
			if !ok {
				vm.bail("IEE: Expected continuation stack and argument")
			}
			// restore stored stack
			wstack, ok := vm.unsafeTupleGet(args, 1).(*data.List)
			if !ok {
				vm.bail("IEE: Expected continuation stack")
			}
			stack := wstack.RawValues()
			vm.stack = append(vm.stack, stack...)
			vm.stackTop = len(vm.stack)
			// push continuation argument
			vm.push(vm.unsafeTupleGet(args, 0))
			// restore registers
			vm.ip = tramp.Ip
			vm.code = tramp.Code
			vm.locals = tramp.Env
		default:
			vm.bail("IEE: cannot handle this call kind")
		}
		break
	}
}

func (vm *Vm) handleEffect(typ *data.Type, arg data.Value) (data.Value, data.Trampoline, bool) {
	for curr := vm.stackTop - 1; curr > -1; curr-- {
		sv := vm.stack[curr]
		if handler, ok := sv.(*data.Handler); ok {
			for ty, h := range handler.Clauses {
				if typ.Equal(ty) {
					len := FUNC_FRAME_SIZE + 1 // handler and function frame
					if data.HandlerCapturesContinuation(h) {
						// does use the continuation so we need to capture whole stack
						// instead of just handler and stack frame
						len = vm.stackTop - curr
					}
					stack := make([]data.Value, len)
					if copied := copy(stack, vm.stack[curr:]); copied != len {
						vm.bail("IEE: Could not copy the stack")
					}
					// throw away other stack frames
					vm.stack = vm.stack[:curr]
					vm.stackTop = curr
					return vm.runHandler(stack, h, arg, handler)
				}
			}
		}
	}
	vm.bail(fmt.Sprintf("Unhandled effect %s with val %s", typ, arg))
	panic("unreachable")
}

func (vm *Vm) runHandler(stack []data.Value, h data.Callable, arg data.Value, handler *data.Handler) (data.Value, data.Trampoline, bool) {
	// first value on the stack is a handler
	// then there is a function frame of function where the
	// handler has been called.
	// We need to retrieve it for the handling function.
	// we pushed                 ip, code, locals
	ip, ok := stack[1].(data.Int)
	if !ok {
		panic("IEE: on return popped value is not an ip")
	}
	code, ok := stack[2].(*data.Code)
	if !ok {
		panic("IEE: on return popped value is not a code")
	}
	env, ok := stack[3].(*data.Env)
	if !ok {
		panic("IEE: on return popped value is not an env")
	}
	sip := vm.ip
	slocals := vm.locals
	scode := vm.code
	vm.ip = ip.Val
	vm.locals = env
	vm.code = code
	if vm.code.Instrs[vm.ip] != isa.PopHandler {
		panic("IEE: expected instruction pointer to be pointing to PopHandler op")
	}
	vm.ip++ // Skip PopHandler instruction as handler is no longer on the stack
	// Tail calls of handlers not supported yet.
	// We would need to know if handler is in tail position.
	// Which we can. We just need to somehow pass it here.
	switch h.Arity() {
	case 1:
		v, t := h.Call(vm, arg)
		return v, t, false
	case 2:
		stack = stack[4:]
		k := data.NewContinuation(stack, handler, sip, scode, slocals)
		v, t := h.Call(vm, arg, k)
		return v, t, false
	default:
		vm.bail("IEE: unsupported arity of effect handling clause")
	}
	panic("Unreachable")
}

func (vm *Vm) popFunctionFrame() (int, *data.Code, *data.Env) {
	env, ok := vm.pop().(*data.Env)
	if !ok {
		panic("IEE: on return popped value is not an env")
	}
	code, ok := vm.pop().(*data.Code)
	if !ok {
		panic("IEE: on return popped value is not a code")
	}
	ip, ok := vm.pop().(data.Int)
	if !ok {
		panic("IEE: on return popped value is not an ip")
	}
	return ip.Val, code, env
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

func (vm *Vm) printStackTrace() {
	fmt.Println("========BACKTRACE========")
	for i := 0; i < vm.stackTop; i++ {
		_, ok := vm.stack[i].(*data.Env)
		if !ok {
			continue
		}
		c, ok := vm.stack[i-1].(*data.Code)
		if !ok {
			continue
		}
		ip, ok := vm.stack[i-2].(data.Int)
		if !ok {
			continue
		}
		vm.printFrame(ip.Val, c)
	}
	fmt.Println("=========================")
}

func (vm *Vm) printFrame(ip int, code *data.Code) {
	if ip >= len(code.Lines) {
		ip = len(code.Lines) - 1
	}
	line := code.Lines[ip] + 1
	r, ok := vm.sources[code.Path]
	var c string
	if !ok {
		c = fmt.Sprintf("Unknown source %v", code.Path)
	} else {
		c = getLine(line, r)
	}
	fmt.Printf("File %s line %d\n\n", code.Path, line)
	fmt.Printf("%s\n", c)
	fmt.Println("---------------------")
}

func (vm *Vm) bail(msg string, args ...interface{}) {
	vm.printStackTrace()
	ip := vm.ip
	if ip >= len(vm.code.Lines) {
		ip = len(vm.code.Lines) - 1
	}
	line := vm.code.Lines[ip] + 1
	r, ok := vm.sources[vm.code.Path]
	var code string
	if !ok {
		code = fmt.Sprintf("Unknown source %v", vm.code.Path)
	} else {
		code = getLine(line, r)
	}
	fmt.Printf("\n\nRuntime error in file %s at line %d\n\n", vm.code.Path, line)
	fmt.Printf(code+"\n\n", args...)
	fmt.Println(msg)
	panic("runtime error")
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
		interner: vm.interner.Clone(),
		// if running mulrithreaded duplicates counts
		gensymc: vm.gensymc,
		// not thread safe
		sources: vm.sources,
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
	default:
		vm.Panic("Unsupported return kind when running closure")
	}
	return data.None
}

func (vm *Vm) unsafeTupleGet(t data.Tuple, i int) data.Value {
	v, err := t.Get(data.NewInt(i))
	if err != nil {
		vm.bail("IEE: tried to take a value from tuple past its indices")
	}
	return v
}

func (vm *Vm) SourceLine() int {
	return vm.code.Lines[vm.ip]
}

func (vm *Vm) FileName() string {
	return vm.code.Path
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
