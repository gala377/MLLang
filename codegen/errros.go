package codegen

import "github.com/gala377/MLLang/syntax/span"

type SourceError interface {
	error
	SourceLoc() span.Span
}
