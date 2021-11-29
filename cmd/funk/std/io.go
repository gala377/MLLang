package std

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/gala377/MLLang/data"
)

// go:embed io.fnk
var funkIo []byte

var ioModule = module{
	Name: "io",
	Entries: map[string]AsValue{
		"print":       &funcEntry{"print", 1, print},
		"printf":      &funcEntry{"printf", 2, printf},
		"printformat": &funcEntry{"printformat", 2, printformat},
	},
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

func printformat(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	fstring, ok := vv[0].(data.String)
	if !ok {
		return nil, errors.New("expected format string as first argument to printformat")
	}
	args, ok := vv[1].(data.Sequence)
	fargs := []interface{}{vv[1].String()}
	if ok {
		fargs = []interface{}{}
		for i := 0; i < args.Len(); i++ {
			v, _ := args.Get(data.NewInt(i)) // cannot fail
			fargs = append(fargs, v.String())
		}
	}
	fmt.Println(fmt.Sprintf(fstring.Val, fargs...))
	return data.None, nil
}
