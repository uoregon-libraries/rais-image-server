package iiif

import (
	"strconv"
	"strings"
)

// SizeType represents the type of scaling which will be performed
type SizeType int

const (
	// STNone is used when the Size struct wasn't able to be parsed form a string
	STNone SizeType = iota
	// STFull means no scaling is requested
	STFull
	// STScaleToWidth requests the image be scaled to a set width (aspect ratio
	// is preserved)
	STScaleToWidth
	// STScaleToHeight requests the image be scaled to a set height (aspect ratio
	// is preserved)
	STScaleToHeight
	// STScalePercent requests the image be scaled by a set percent of its size
	// (aspect ratio is preserved)
	STScalePercent
	// STExact requests the image be resized to precise width and height
	// dimensions (aspect ratio is not preserved)
	STExact
	// STBestFit requests the image be resized *near* the given width and height
	// dimensions (aspect ratio is preserved)
	STBestFit
)

// Size represents the type of scaling as well as the parameters for scaling
// for an IIIF 2.0 server
type Size struct {
	Type    SizeType
	Percent float64
	W, H    int
}

// StringToSize creates a Size from a string as seen in an IIIF URL.
func StringToSize(p string) Size {
	if p == "full" {
		return Size{Type: STFull}
	}

	s := Size{Type: STNone}

	if len(p) > 4 && p[0:4] == "pct:" {
		s.Type = STScalePercent
		s.Percent, _ = strconv.ParseFloat(p[4:], 64)
		return s
	}

	if p[0:1] == "!" {
		s.Type = STBestFit
		p = p[1:]
	}

	vals := strings.Split(p, ",")
	s.W, _ = strconv.Atoi(vals[0])
	s.H, _ = strconv.Atoi(vals[1])

	if s.Type == STNone {
		if vals[0] == "" {
			s.Type = STScaleToHeight
		} else if vals[1] == "" {
			s.Type = STScaleToWidth
		} else {
			s.Type = STExact
		}
	}

	return s
}

// Valid returns whether the size has a valid type, and if so, whether the
// parameters are valid for that type
func (s Size) Valid() bool {
	switch s.Type {
	case STFull:
		return true
	case STScaleToWidth:
		return s.W > 0
	case STScaleToHeight:
		return s.H > 0
	case STScalePercent:
		return s.Percent > 0
	case STExact, STBestFit:
		return s.W > 0 && s.H > 0
	}

	return false
}
