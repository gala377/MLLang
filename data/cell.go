package data

import "fmt"

type Cell struct {
	Val Value
}

func (c *Cell) Equal(o Value) bool {
	if oc, ok := o.(*Cell); ok {
		return c.Val.Equal(oc.Val)
	}
	return false
}

func (c *Cell) String() string {
	return fmt.Sprintf("<cell(%s)>", c.Val)
}

func (c *Cell) Set(v Value) {
	c.Val = v
}

func (c *Cell) Get() Value {
	return c.Val
}
