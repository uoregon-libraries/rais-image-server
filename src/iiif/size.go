package iiif

import (
	"image"
	"math"
	"strconv"
	"strings"
)

// SizeType represents the type of scaling which will be performed
type SizeType int

const (
	// STNone is used when the Size struct wasn't able to be parsed form a string
	STNone SizeType = iota
	// STMax requests the maximum size the server supports
	STMax
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

// Size represents the type of scaling as well as the parameters for scaling.
// Upscale is set when a IIIF 3.0 request uses the "^" prefix to request an image
// larger than the extracted region.
type Size struct {
	Type    SizeType
	Percent float64
	W, H    int
	Upscale bool
}

// StringToSize creates a Size from a string as seen in a IIIF URL.  Parsing is
// version-neutral: both "full" (v2) and "max" (v2/v3) yield a value, and the "^"
// upscaling prefix (v3) is recorded in Upscale.  Whether a given form is legal
// for a particular spec version is enforced at the URL level.
func StringToSize(p string) Size {
	if p == "" {
		return Size{}
	}

	// The v3 upscaling prefix may precede any of the other size forms
	var upscale bool
	if len(p) > 0 && p[0:1] == "^" {
		upscale = true
		p = p[1:]
	}

	if p == "full" {
		return Size{Type: STFull, Upscale: upscale}
	}
	if p == "max" {
		return Size{Type: STMax, Upscale: upscale}
	}

	s := Size{Type: STNone, Upscale: upscale}

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
	if len(vals) != 2 {
		return s
	}
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
	case STFull, STMax:
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

// GetResize determines how a given region would be resized and returns a
// rectangle representing the scaled image's dimensions.  If STMax is in use,
// this returns the full region, as only the image server itself would know its
// capabilities and therefore it shouldn't call this in that scenario.
func (s Size) GetResize(region image.Rectangle) image.Rectangle {
	w, h := region.Dx(), region.Dy()

	var cloned = s
	switch s.Type {
	case STScaleToWidth:
		cloned.H = math.MaxInt32
		w, h = cloned.getBestFit(w, h)
	case STScaleToHeight:
		cloned.W = math.MaxInt32
		w, h = cloned.getBestFit(w, h)
	case STExact:
		w, h = cloned.W, cloned.H
	case STBestFit:
		w, h = cloned.getBestFit(w, h)
	case STScalePercent:
		w = int(float64(w) * cloned.Percent / 100.0)
		h = int(float64(h) * cloned.Percent / 100.0)
	}

	return image.Rect(0, 0, w, h)
}

// getBestFit preserves the aspect ratio while determining the proper scaling
// factor to get width and height adjusted to fit within the width and height
// of the desired size operation
func (s Size) getBestFit(w, h int) (width int, height int) {
	fW, fH, fsW, fsH := float64(w), float64(h), float64(s.W), float64(s.H)
	sf := fsW / fW
	if sf*fH > fsH {
		sf = fsH / fH
	}
	return int(sf * fW), int(sf * fH)
}
