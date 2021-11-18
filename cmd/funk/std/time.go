package std

import (
	"errors"
	"time"

	"github.com/gala377/MLLang/data"
)

var timeModule = module{
	Name: "time",
	Entries: map[string]AsValue{
		"sleep": &funcEntry{"sleep", 1, wait},
		"now":   &funcEntry{"now", 0, now},
	},
}

func wait(vm data.VmProxy, vv ...data.Value) (data.Value, error) {
	arg, ok := vv[0].(data.Int)
	if !ok {
		return nil, errors.New("wait expects int as a measure of time to wait")
	}
	time.Sleep(time.Duration(arg.Val) * time.Second)
	return data.None, nil
}

func now(_ data.VmProxy, vv ...data.Value) (data.Value, error) {
	n := time.Now()
	return data.NewInt(int(n.Unix())), nil
}
