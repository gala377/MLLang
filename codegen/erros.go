package codegen

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/gala377/MLLang/syntax/span"
)

type SourceError interface {
	error
	SourceLoc() span.Span
}

func printWithSourceExact(source io.ReaderAt, srcerr SourceError) {
	loc := srcerr.SourceLoc()
	code, err := loc.Extract(source)
	if err != nil {
		log.Panicf("unexpected error when extracting source code %s", err)
	}
	line, col := loc.Beg.Line, loc.Beg.Column
	log.Printf("Error at line %d, column %d\n\n", line, col)
	log.Printf("%s\n\n", code)
	log.Println(srcerr.Error())
}

func printWithSourceLine(source *bytes.Reader, srcerr SourceError) {
	source.Seek(0, 0)
	loc := srcerr.SourceLoc()
	line, col := loc.Beg.Line, loc.Beg.Column
	eline := loc.End.Line
	s := bufio.NewScanner(source)
	s.Split(bufio.ScanLines)
	var text []string
	for i := uint(0); s.Scan(); i++ {
		if i >= line {
			text = append(text, s.Text())
		}
		if i >= eline {
			break
		}
	}
	code := strings.Join(text, "\n")
	fmt.Printf("Error at line %d, column %d\n", line+1, col)
	fmt.Printf("%s\n\n", code)
	fmt.Println(srcerr.Error())
}

var PrintWithSource = printWithSourceLine
