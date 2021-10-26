package codegen

import (
	"fmt"
	"io"

	"github.com/gala377/MLLang/syntax/span"
)

type SourceError interface {
	error
	SourceLoc() span.Span
}

func PrintWithSource(source io.ReaderAt, srcerr SourceError) {
	loc := srcerr.SourceLoc()
	code, err := loc.Extract(source)
	if err != nil {
		panic(fmt.Sprintf("unexpected error when extracting source code %s", err))
	}
	line, col := loc.Beg.Line, loc.Beg.Column
	fmt.Printf("Error at line %d, column %d\n\n", line, col)
	fmt.Println(string(code))
	fmt.Println(srcerr.Error())
}
