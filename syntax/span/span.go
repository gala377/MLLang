package span

// Span is represents a fragment of source that the given
// production is associated with. The fields correspond to
// the byte offset from the begging of the source reader.
type Span struct {
	Beg uint
	End uint
}
