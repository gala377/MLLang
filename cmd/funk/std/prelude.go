package std

import (
	"errors"
	"fmt"

	_ "embed"

	"github.com/gala377/MLLang/data"
)

//go:embed prelude.fnk
var funkPrelude []byte

//go:embed struct.fnk
var funkStruct []byte

func add(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	a1, a2 := vv[0], vv[1]
	i1, ok := a1.(data.Number)
	if !ok {
		return nil, fmt.Errorf("can only add numbers")
	}
	i2, ok := a2.(data.Number)
	if !ok {
		return nil, fmt.Errorf("can only add numbers")
	}
	return data.CallNumOp(i1, data.AddOp, i2), nil
}

func sub(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	a1, a2 := vv[0], vv[1]
	i1, ok := a1.(data.Number)
	if !ok {
		return nil, fmt.Errorf("can only sub numbers")
	}
	i2, ok := a2.(data.Number)
	if !ok {
		return nil, fmt.Errorf("can only sub numbers")
	}
	return data.CallNumOp(i1, data.SubOpp, i2), nil
}

func mul(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	a1, a2 := vv[0], vv[1]
	i1, ok := a1.(data.Number)
	if !ok {
		return nil, fmt.Errorf("can only mul numbers")
	}
	i2, ok := a2.(data.Number)
	if !ok {
		return nil, fmt.Errorf("can only mul numbers")
	}
	return data.CallNumOp(i1, data.MulOp, i2), nil
}

func div(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	a1, a2 := vv[0], vv[1]
	i1, ok := a1.(data.Number)
	if !ok {
		return nil, fmt.Errorf("can only div numbers")
	}
	i2, ok := a2.(data.Number)
	if !ok {
		return nil, fmt.Errorf("can only div numbers")
	}
	return data.CallNumOp(i1, data.DevideOp, i2), nil
}

func neg(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	i1, ok := vv[0].(data.Number)
	if !ok {
		return nil, fmt.Errorf("can only negate numbers")
	}
	return i1.Neg(), nil
}

func printf(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	format, val := vv[0], vv[1]
	sfmt, ok := format.(data.String)
	if !ok {
		return nil, fmt.Errorf("first argument to printf has to be a string")
	}
	fmt.Printf(sfmt.Val+"\n", val)
	return data.None, nil
}

func print(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	val := vv[0]
	msg := val.String()
	if s, ok := val.(data.String); ok {
		msg = s.Val
	}
	fmt.Println(msg)
	return data.None, nil
}
func lessThan(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	a, b := vv[0], vv[1]
	ai, ok := a.(data.Int)
	if !ok {
		return nil, errors.New("lessThan can only be called on integers")
	}
	bi, ok := b.(data.Int)
	if !ok {
		return nil, errors.New("lessThan can only be called on integers")
	}
	ret := data.NewBool(ai.Val < bi.Val)
	return ret, nil
}

func equal(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	a, b := vv[0], vv[1]
	return data.NewBool(a.Equal(b)), nil
}

func not(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	v, ok := vv[0].(data.Bool)
	if !ok {
		return nil, fmt.Errorf("not expects a bool got %s", vv[0])
	}
	return data.NewBool(!v.Val), nil
}

func vmPanic(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	vm.Panic(vv[0].String())
	return nil, nil
}

func vmGenSym(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	return vm.GenerateSymbol(), nil
}

func vmSpawn(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	c, ok := vv[0].(data.Callable)
	if !ok || c.Arity() > 0 {
		return nil, errors.New("spawn expects a callable of arity 0 to run")
	}
	cloned := vm.Clone()
	go cloned.RunClosure(c)
	return data.None, nil
}

func isInt(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	_, ok := vv[0].(data.Int)
	return data.NewBool(ok), nil
}

func isList(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	_, ok := vv[0].(*data.List)
	return data.NewBool(ok), nil
}
func isTuple(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	_, ok := vv[0].(*data.Tuple)
	return data.NewBool(ok), nil
}

func isSeq(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	_, ok := vv[0].(data.Sequence)
	return data.NewBool(ok), nil
}

func isString(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	_, ok := vv[0].(data.String)
	return data.NewBool(ok), nil
}

func isSymbol(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	_, ok := vv[0].(data.Symbol)
	return data.NewBool(ok), nil
}

func isFunction(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	_, ok := vv[0].(data.Callable)
	return data.NewBool(ok), nil
}

func isFloat(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	_, ok := vv[0].(data.Float)
	return data.NewBool(ok), nil
}

func isBool(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	_, ok := vv[0].(data.Bool)
	return data.NewBool(ok), nil
}

func isRecord(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	_, ok := vv[0].(*data.Record)
	return data.NewBool(ok), nil
}

func boolAnd(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	a, b := vv[0], vv[1]
	ab, ok := a.(data.Bool)
	if !ok {
		return nil, errors.New("and only works on booleans")
	}
	bb, ok := b.(data.Bool)
	if !ok {
		return nil, errors.New("and only works on booleans")
	}
	return data.NewBool(ab.Val && bb.Val), nil
}

func boolOr(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	a, b := vv[0], vv[1]
	ab, ok := a.(data.Bool)
	if !ok {
		return nil, errors.New("or only works on booleans")
	}
	bb, ok := b.(data.Bool)
	if !ok {
		return nil, errors.New("or only works on booleans")
	}
	return data.NewBool(ab.Val && bb.Val), nil
}
