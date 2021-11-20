package data

type (
	NumKind = int

	Number interface {
		Cast(NumKind) Number
		Kind() NumKind
		Neg() Value
	}

	NumOp = [kindcount]func(Number, Number) Value
)

const (
	IntKind NumKind = iota
	FloatKind
	kindcount
)

var AddOp = NumOp{
	IntKind: func(n1, n2 Number) Value {
		i1, i2 := n1.(Int), n2.(Int)
		return NewInt(i1.Val + i2.Val)
	},
	FloatKind: func(n1, n2 Number) Value {
		i1, i2 := n1.(Float), n2.(Float)
		return NewFloat(i1.Val + i2.Val)
	},
}

var SubOpp = NumOp{
	IntKind: func(n1, n2 Number) Value {
		i1, i2 := n1.(Int), n2.(Int)
		return NewInt(i1.Val - i2.Val)
	},
	FloatKind: func(n1, n2 Number) Value {
		i1, i2 := n1.(Float), n2.(Float)
		return NewFloat(i1.Val - i2.Val)
	},
}

var MulOp = NumOp{
	IntKind: func(n1, n2 Number) Value {
		i1, i2 := n1.(Int), n2.(Int)
		return NewInt(i1.Val * i2.Val)
	},
	FloatKind: func(n1, n2 Number) Value {
		i1, i2 := n1.(Float), n2.(Float)
		return NewFloat(i1.Val * i2.Val)
	},
}

var DevideOp = NumOp{
	IntKind: func(n1, n2 Number) Value {
		i1, i2 := n1.(Int), n2.(Int)
		return NewInt(i1.Val / i2.Val)
	},
	FloatKind: func(n1, n2 Number) Value {
		i1, i2 := n1.(Float), n2.(Float)
		return NewFloat(i1.Val / i2.Val)
	},
}

func CallNumOp(a Number, op NumOp, b Number) Value {
	nkind := PromoteNumbers(a.Kind(), b.Kind())
	ap, bp := a.Cast(nkind), b.Cast(nkind)
	return op[nkind](ap, bp)
}

func PromoteNumbers(a NumKind, b NumKind) NumKind {
	switch a {
	case IntKind:
		switch b {
		case IntKind:
			return IntKind
		case FloatKind:
			return FloatKind
		}
	case FloatKind:
		return FloatKind
	}
	return FloatKind
}
