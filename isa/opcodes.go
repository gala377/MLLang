package isa

type Op = byte

const (
	Return Op = iota
	Constant
	Constant2
	Call
	Jump
	JumpIfFalse
	LoadDyn
	StoreDyn
	StoreLocal
	LoadLocal
	Pop
	DefGlobal
	DefLocal
	Lambda
	PushNone
)
