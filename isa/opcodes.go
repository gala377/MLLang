package isa

type Op = byte

const (
	Return Op = iota
	Constant
	Call
	Jump
	JumpIf
)
