package std

import (
	"fmt"

	"github.com/gala377/MLLang/data"
)

var inspectModule = module{
	Name: "inspect",
	Entries: map[string]AsValue{
		"arity": &funcEntry{"arity", 1, inspectArity},
	},
}

func inspectArity(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	arg, ok := vv[0].(data.Callable)
	if !ok {
		return nil, fmt.Errorf("expected callable got %v", arg)
	}
	return data.NewInt(arg.Arity()), nil
}
