package span

import "fmt"

// Span is represents a fragment of source that the given
// production is associated with. The fields correspond to
// the byte offset from the begging of the source reader.
type Span struct {
	Beg uint
	End uint
}

func NewSpan(beg, end uint) Span {
	if beg > end {
		err := fmt.Sprintf("wrong span creation arguments: assertion beg <= end, actual beg=%v end=%v", beg, end)
		panic(err)
	}
	return Span{beg, end}
}
