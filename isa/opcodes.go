package isa

type Op = byte

const (
	Return Op = iota
	Constant
	Constant2
	Call
	Call0
	Call1
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
	CResume
	CCapture
	CPrompt
	InstallHandler
	PopHandler
	PerformEffect
)
