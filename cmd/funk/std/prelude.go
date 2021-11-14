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

func seqGet(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	s, i := vv[0], vv[1]
	as, ok := s.(data.Sequence)
	if !ok {
		return nil, errors.New("get can only be called on sequences")
	}
	idx, ok := i.(data.Int)
	if !ok {
		return nil, errors.New("index of get has to be an integer")
	}
	return as.Get(idx)
}

func seqSet(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	s, i, v := vv[0], vv[1], vv[2]
	as, ok := s.(data.MutableSequence)
	if !ok {
		return nil, errors.New("get can only be called on mutable sequences")
	}
	idx, ok := i.(data.Int)
	if !ok {
		return nil, errors.New("index of set has to be an integer")
	}
	return data.None, as.Set(idx, v)
}

func seqLen(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	s := vv[0]
	as, ok := s.(data.Sequence)
	if !ok {
		return nil, errors.New("length can only be called on sequences")
	}
	return data.NewInt(as.Len()), nil
}

func seqAppend(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	s, v := vv[0], vv[1]
	as, ok := s.(data.Appendable)
	if !ok {
		return nil, errors.New("append can only be called on appendable sequences")
	}
	return s, as.Append(v)
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
