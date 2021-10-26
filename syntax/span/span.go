package span

import (
	"fmt"
	"io"
)

// Span is represents a fragment of source that the given
// production is associated with. The fields correspond to
// the byte offset from the begging of the source reader.

type Position struct {
	Line   uint
	Column uint
	Offset uint
}
type Span struct {
	Beg Position
	End Position
}

func NewSpan(beg, end Position) Span {
	if beg.Offset > end.Offset {
		err := fmt.Sprintf("wrong span creation arguments: assertion beg <= end, actual beg=%v end=%v", beg, end)
		panic(err)
	}
	return Span{beg, end}
}

func (s Span) Extract(source io.ReaderAt) ([]byte, error) {
	begoff := s.Beg.Offset
	readbytes := s.End.Offset - s.Beg.Offset
	out := make([]byte, readbytes)
	n, err := source.ReadAt(out, int64(begoff))
	if n != int(readbytes) {
		panic("Extract read less bytes than it expected, this should never happen")
	}
	return out, err

}
