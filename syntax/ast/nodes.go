package ast

import (
	"github.com/gala377/MLLang/syntax/span"
)

type (
	Node interface {
		BegPos() span.Position
		EndPos() span.Position
	}
)
