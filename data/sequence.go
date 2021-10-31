package data

type (
	Sequence interface {
		Get(index *Int) (Value, error)
	}

	MutableSequence interface {
		Sequence
		Set(index *Int, v Value) error
	}

	Appendable interface {
		Append(v Value) error
	}
)
