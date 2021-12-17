package codegen

import (
	"bytes"
	"fmt"

	"github.com/gala377/MLLang/data"
	"github.com/gala377/MLLang/syntax"
)

func Compile(path string, source []byte, interner *Interner) (*data.Code, error) {
	sr := bytes.NewReader(source)
	p := syntax.NewParser(sr)
	ast := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Print("Parsing error:")
		for _, e := range p.Errors() {
			PrintWithSource(path, sr, e)
		}
		return nil, fmt.Errorf("syntax errors")
	}
	e := NewEmitter(path, interner)
	c, errs := e.Compile(ast)
	if len(errs) > 0 {
		fmt.Print("Compilation errors:\n")
		for _, e := range errs {
			PrintWithSource(path, sr, e)
		}
		return nil, fmt.Errorf("compilation errors")
	}
	fmt.Printf("Compiling source for %v\n", path)
	c.Path = path
	return c, nil
}
