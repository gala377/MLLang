package span

import (
	"strings"
	"testing"
)

func spanFromOffsets(beg, end uint) *Span {
	return &Span{
		Beg: Position{
			Offset: beg,
		},
		End: Position{
			Offset: end,
		},
	}
}

func TestExtract(t *testing.T) {
	source := strings.NewReader("aabbcc")
	s := spanFromOffsets(2, 4)
	want := "bb"
	r, err := s.Extract(source)
	got := string(r)
	if err != nil {
		t.Errorf("Extract returned an error %s", err)
		return
	}
	if got != want {
		t.Errorf("want=\"%s\" got=\"%s\"", want, got)
	}
}
