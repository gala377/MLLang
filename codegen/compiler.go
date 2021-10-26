package codegen

import (
	"bytes"
	"fmt"

	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/syntax"
)

func Compile(source []byte, interner *Interner) (*data.Code, error) {
	sr := bytes.NewReader(source)
	p := syntax.NewParser(sr)
	ast := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Print("Parsing error:")
		for _, e := range p.Errors() {
			PrintWithSource(sr, e)
		}
		return nil, fmt.Errorf("syntax errors")
	}
	e := NewEmitter(interner)
	c, errs := e.Compile(ast)
	if len(errs) > 0 {
		fmt.Print("Compilation errors:\n")
		for _, e := range errs {
			PrintWithSource(sr, e)
		}
		return nil, fmt.Errorf("compilation errors")
	}
	return c, nil
}
