package std

import (
	_ "embed"
	"fmt"

	"github.com/gala377/MLLang/data"
)

func toString(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	return data.String{Val: vv[0].String()}, nil
}

func toSymbol(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	switch v := vv[0].(type) {
	case data.Symbol:
		return v, nil
	case data.String:
		return vm.CreateSymbol(v.Val), nil
	default:
		return nil, fmt.Errorf("toSymbol expects strings or symbols got %s", v)
	}
}

var convModule module = module{
	Name: "conv",
	Entries: map[string]AsValue{
		"toString": &funcEntry{"toString", 1, toString},
		"toSymbol": &funcEntry{"toSymbol", 1, toSymbol},
	},
}

//go:embed conv.fnk
var funkConv []byte
