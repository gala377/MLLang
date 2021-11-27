package data

import "fmt"

// Label can be used as a marker on the stack
type Label struct{}

func NewLable() *Label {
	return &Label{}
}

func (l *Label) String() string {
	return fmt.Sprintf("label %d", l)
}

func (l *Label) Equal(o Value) bool {
	ol, ok := o.(*Label)
	return ok && ol == l
}
