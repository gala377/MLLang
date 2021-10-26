package main

import (
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
	sfmt, ok := val.(data.String)
	if !ok {
		return nil, fmt.Errorf("first argument to print has to be a string")
	}
	fmt.Print(sfmt.Val + "\n")
	return data.None, nil
}

func toString(vv ...data.Value) (data.Value, error) {
	return data.String{Val: vv[0].String()}, nil
}
