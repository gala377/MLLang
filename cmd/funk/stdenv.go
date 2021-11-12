package main

import (
	"errors"
	"fmt"

	"github.com/gala377/MLLang/data"
)

type funcEntry struct {
	name  string
	arity int
	f     func(...data.Value) (data.Value, error)
}

type StandardEnv = []funcEntry

var stdEnv = StandardEnv{
	{"add", 2, add},
	{"printf", 2, printf},
	{"print", 1, print},
	{"toString", 1, toString},
	{"append", 2, seqAppend},
	{"get", 2, seqGet},
	{"set", 3, seqSet},
	{"lessThan?", 2, lessThan},
}

func add(vv ...data.Value) (data.Value, error) {
	a1, a2 := vv[0], vv[1]
	i1, ok := a1.(*data.Int)
	if !ok {
		return nil, fmt.Errorf("can only add integers")
	}
	i2, ok := a2.(*data.Int)
	if !ok {
		return nil, fmt.Errorf("can only add integers")
	}
	return data.NewInt(i1.Val + i2.Val), nil
}

func printf(vv ...data.Value) (data.Value, error) {
	format, val := vv[0], vv[1]
	sfmt, ok := format.(data.String)
	if !ok {
		return nil, fmt.Errorf("first argument to printf has to be a string")
	}
	fmt.Printf(sfmt.Val+"\n", val)
	return data.None, nil
}

func print(vv ...data.Value) (data.Value, error) {
	val := vv[0]
	msg := val.String()
	if s, ok := val.(data.String); ok {
		msg = s.Val
	}
	fmt.Println(msg)
	return data.None, nil
}

func toString(vv ...data.Value) (data.Value, error) {
	return data.String{Val: vv[0].String()}, nil
}

func seqGet(vv ...data.Value) (data.Value, error) {
	s, i := vv[0], vv[1]
	as, ok := s.(data.Sequence)
	if !ok {
		return nil, errors.New("get can only be called on sequences")
	}
	idx, ok := i.(*data.Int)
	if !ok {
		return nil, errors.New("index of get has to be an integer")
	}
	return as.Get(idx)
}

func seqSet(vv ...data.Value) (data.Value, error) {
	s, i, v := vv[0], vv[1], vv[2]
	as, ok := s.(data.MutableSequence)
	if !ok {
		return nil, errors.New("get can only be called on mutable sequences")
	}
	idx, ok := i.(*data.Int)
	if !ok {
		return nil, errors.New("index of set has to be an integer")
	}
	return data.None, as.Set(idx, v)
}

func seqAppend(vv ...data.Value) (data.Value, error) {
	s, v := vv[0], vv[1]
	as, ok := s.(data.Appendable)
	if !ok {
		return nil, errors.New("append can only be called on appendable sequences")
	}
	return data.None, as.Append(v)
}

func lessThan(vv ...data.Value) (data.Value, error) {
	a, b := vv[0], vv[1]
	ai, ok := a.(*data.Int)
	if !ok {
		return nil, errors.New("lessThan can only be called on integers")
	}
	bi, ok := b.(*data.Int)
	if !ok {
		return nil, errors.New("lessThan can only be called on integers")
	}
	ret := data.NewBool(ai.Val < bi.Val)
	return ret, nil
}
