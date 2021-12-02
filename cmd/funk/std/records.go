package std

import (
	_ "embed"
	"errors"

	"github.com/gala377/MLLang/data"
)

//go:embed records.fnk
var funkRecords []byte

var recordsModule = module{
	Name: "records",
	Entries: map[string]AsValue{
		"getField":  &funcEntry{"getField", 3, getField},
		"hasField?": &funcEntry{"hasField?", 2, hasField},
	},
}

func getField(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	rec, ok := vv[0].(*data.Record)
	if !ok {
		return nil, errors.New("first argument to getField should be a record")
	}
	field, ok := vv[1].(data.Symbol)
	if !ok {
		return nil, errors.New("second argument to getField should be a symbol")
	}
	def := vv[2]
	v, ok := rec.GetField(field)
	if !ok {
		return def, nil
	}
	return v, nil
}

func hasField(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	rec, ok := vv[0].(*data.Record)
	if !ok {
		return nil, errors.New("first argument to hasField should be a record")
	}
	field, ok := vv[1].(data.Symbol)
	if !ok {
		return nil, errors.New("second argument to hasField should be a symbol")
	}
	_, ok = rec.GetField(field)
	return data.NewBool(ok), nil
}
