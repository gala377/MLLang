package std

import (
	"errors"
	"fmt"

	_ "embed"

	"github.com/gala377/MLLang/data"
)

//go:embed prelude.fnk
var funkPrelude []byte

func add(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	a1, a2 := vv[0], vv[1]
	i1, ok := a1.(data.Int)
	if !ok {
		return nil, fmt.Errorf("can only add integers")
	}
	i2, ok := a2.(data.Int)
	if !ok {
		return nil, fmt.Errorf("can only add integers")
	}
	return data.NewInt(i1.Val + i2.Val), nil
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
