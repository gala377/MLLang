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
	Lambda
	PushNone
	MakeCell
	MakeList
	MakeTuple
)
