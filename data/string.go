package data

type String struct {
	val string
}

func NewString(v string) String {
	return String{v}
}

func (s String) String() string {
	return s.val
}

func (s String) Equal(o Value) bool {
	if os, ok := o.(String); ok {
		return s.val == os.val
	}
	return false
}
