package main

import (
	"fmt"

	"github.com/gala377/MLLang/data"
)

type funcEntry struct {
	name  string
	arity int
	f     func(...data.Value) data.Value
}

type StandardEnv = []funcEntry

var stdEnv = StandardEnv{
	{"add", 2, add},
	{"printf", 2, printf},
	{"print", 1, print},
}

func add(vv ...data.Value) data.Value {
	a1, a2 := vv[0], vv[1]
	i1, ok := a1.(*data.Int)
	if !ok {
		panic("Can only add integers")
	}
	i2, ok := a2.(*data.Int)
	if !ok {
		panic("Can only add integers")
	}
	return data.NewInt(i1.Val + i2.Val)
}

func printf(vv ...data.Value) data.Value {
	format, val := vv[0], vv[1]
	sfmt, ok := format.(data.String)
	if !ok {
		panic("First argument to printf has to be a string")
	}
	fmt.Printf(sfmt.Val+"\n", val)
	return data.NewInt(0)
}

func print(vv ...data.Value) data.Value {
	val := vv[0]
	sfmt, ok := val.(data.String)
	if !ok {
		panic("First argument to print has to be a string")
	}
	fmt.Print(sfmt.Val + "\n")
	return data.NewInt(0)
}
