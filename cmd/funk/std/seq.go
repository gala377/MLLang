package std

import (
	_ "embed"
	"errors"

	"github.com/gala377/MLLang/data"
)

//go:embed seq.fnk
var funkSeq []byte

var seqModule = module{
	Name: "seq",
	Entries: map[string]AsValue{
		"get":    &funcEntry{"get", 2, seqGet},
		"set":    &funcEntry{"set", 3, seqSet},
		"append": &funcEntry{"append", 2, seqAppend},
		"len":    &funcEntry{"len", 1, seqLen},
	},
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
