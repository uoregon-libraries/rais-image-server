package img

// Constraint holds maximums the server is willing to return in image dimensions
type Constraint struct {
	Width  int
	Height int
	Area   int64
}

// SmallerThanAny returns true if the constraint's maximums are exceeded by the
// given width and height
func (c Constraint) SmallerThanAny(w, h int) bool {
	return w > c.Width || h > c.Height || int64(w)*int64(h) > c.Area
}
