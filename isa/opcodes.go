package isa

type Op = byte

const (
	Return Op = iota
	Constant
	Constant2
	// Calls function with n arguments
	// where n is this instructions argument.
	// Pops n+1 values from the stack and pushes 1
	Call
	// Specialised call. Has no arguments.
	// Calls function on top of the stack with no arguments.
	Call0
	// Specialised call. Has no arguments.
	// Calls function with one argument.
	Call1
	Jump
	JumpIfFalse
	JumpBack
	// Loads from global environment.
	LoadDyn
	// Stores value in the global environemtn
	StoreDyn
	// Stores value into a cell
	LoadDeref
	// Loads value from a cell
	StoreDeref
	StoreLocal
	LoadLocal
	Pop
	Rotate
	DefGlobal
	DefLocal
	Closure
	PushNone
	MakeCell
	MakeList
	MakeTuple
	MakeRecord
	MakeEffect
	GetField
	SetField
	// Creates the handler from values on the stack.
	// Then pushes it.
	InstallHandler
	// Pops two values from the stack.
	// Expects second one to be an effect handler.
	// Then pushes the first value again on the stack
	PopHandler
	PerformEffect
	// Inspects a continuation on top of the stack
	// and installs the handler that the continuation
	// has been created under
	Resume
)
