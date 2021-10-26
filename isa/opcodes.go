package isa

type Op = byte

const (
	Return Op = iota
	Constant
	Constant2
	Call
	Jump
	JumpIfFalse
	DynLookup
	LocalLookup
	Pop
	DefGlobal
	DefLocal
	Lambda
)
