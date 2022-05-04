package std

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gala377/MLLang/data"
)

var stringsModule = module{
	Name: "strings",
	Entries: map[string]AsValue{
		"fmt":  &funcEntry{"fmt", 2, format},
		"join": &funcEntry{"join", 2, join},
	},
}

func format(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	fmts, ok := vv[0].(data.String)
	if !ok {
		return nil, errors.New("first argument to strings.fmt should be a string")
	}
	args, ok := vv[1].(data.Sequence)
	if !ok {
		return nil, errors.New("second argument to strings.fmt should be a sequence")
	}
	argc := args.Len()
	fargs := make([]interface{}, 0, argc)
	for i := 0; i < argc; i++ {
		arg, _ := args.Get(data.NewInt(i))
		if asstr, ok := arg.(data.String); ok {
			fargs = append(fargs, asstr.Val)
		} else {
			fargs = append(fargs, arg.String())
		}
	}
	res := fmt.Sprintf(fmts.Val, fargs...)
	return data.NewString(res), nil
}

func join(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	with, ok := vv[0].(data.String)
	if !ok {
		return nil, errors.New("first argument to strings.join should be a string")
	}
	seq, ok := vv[1].(data.Sequence)
	if !ok {
		return nil, errors.New("second argument to strings.join should be a sequence")
	}
	argc := seq.Len()
	ss := make([]string, 0, argc)
	for i := 0; i < argc; i++ {
		arg, _ := seq.Get(data.NewInt(i))
		s, ok := arg.(data.String)
		if !ok {
			return nil, errors.New("strings.join accepts list of strings")
		}
		ss = append(ss, s.Val)
	}
	return data.NewString(strings.Join(ss, with.Val)), nil
}
