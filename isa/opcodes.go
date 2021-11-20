package isa

type Op = byte

const (
	Return Op = iota
	Constant
	Constant2
	Call
	Jump
	JumpIfFalse
	JumpBack
	LoadDyn
	StoreDyn
	LoadDeref
	StoreDeref
	StoreLocal
	LoadLocal
	Pop
	DefGlobal
	DefLocal
	Closure
	PushNone
	MakeCell
	MakeList
	MakeTuple
	MakeRecord
	GetField
	SetField
	ResumeContinuation
	CaptureContinuation
)
