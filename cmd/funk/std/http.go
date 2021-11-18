package std

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gala377/MLLang/data"
)

var httpModule = module{
	Name: "http",
	Entries: map[string]AsValue{
		"addHandler": &funcEntry{"addHandler", 2, httpHandle},
		"serve":      &funcEntry{"serve", 1, httpServe},
	},
}

func httpHandle(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	path, ok := vv[0].(data.String)
	if !ok {
		return nil, errors.New("http.handle's first argument should be an endpoints path")
	}
	f, ok := vv[1].(data.Callable)
	if !ok || f.Arity() != 1 {
		return nil, errors.New("http.handle expects function of arity 1 that takse a request")
	}
	handlervm := vm.Clone()
	handler := func(w http.ResponseWriter, r *http.Request) {
		res := handlervm.RunClosure(f, data.None)
		fmt.Fprint(w, res.String())
	}
	http.HandleFunc(path.Val, handler)
	return data.None, nil
}

func httpServe(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	address, ok := vv[0].(data.String)
	if !ok {
		return nil, errors.New("serve expects a server address as a string")
	}
	http.ListenAndServe(address.Val, nil)
	return data.None, nil
}
