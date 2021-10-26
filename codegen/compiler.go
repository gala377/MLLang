package codegen

import (
	"bytes"
	"fmt"

	"github.com/gala377/MLLang/isa"
	"github.com/gala377/MLLang/syntax"
)

func Compile(source []byte) (*isa.Code, error) {
	p := syntax.NewParser(bytes.NewReader(source))
	ast := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Print("Parsing error:")
		for _, e := range p.Errors() {
			fmt.Print(e)
		}
		return nil, fmt.Errorf("Syntax errors")
	}
	e := NewEmitter()
	c, errs := e.Compile(ast)
	if len(errs) > 0 {
		fmt.Print("Compilation errors:")
		for _, e := range p.Errors() {
			fmt.Print(e)
		}
		return nil, fmt.Errorf("Compilation errors")
	}
	return c, nil
}
