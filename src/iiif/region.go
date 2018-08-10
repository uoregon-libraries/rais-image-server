package iiif

import (
	"strconv"
	"strings"
)

// A RegionType tells us what a Region is representing so we know how to apply
// the x/y/w/h values
type RegionType int

const (
	// RTNone means we didn't find a valid region string
	RTNone RegionType = iota
	// RTFull means we ignore x/y/w/h and use the whole image
	RTFull
	// RTPercent means we interpret x/y/w/h as percentages of the image size
	RTPercent
	// RTPixel means we interpret x/y/w/h as precise coordinates within the image
	RTPixel
	// RTSquare means a square region where w/h are the image's shortest dimension
	RTSquare
)

// Region represents the part of the image we'll manipulate.  It can be thought
// of as the cropping rectangle.
type Region struct {
	Type       RegionType
	X, Y, W, H float64
}

// StringToRegion takes a string representing a region, as seen in an IIIF URL,
// and fills in the values based on the string's format.
func StringToRegion(p string) Region {
	if p == "full" {
		return Region{Type: RTFull}
	}
	if p == "square" {
		return Region{Type: RTSquare}
	}

	r := Region{Type: RTPixel}
	if len(p) > 4 && p[0:4] == "pct:" {
		r.Type = RTPercent
		p = p[4:]
	}

	vals := strings.Split(p, ",")
	r.X, _ = strconv.ParseFloat(vals[0], 64)
	r.Y, _ = strconv.ParseFloat(vals[1], 64)
	r.W, _ = strconv.ParseFloat(vals[2], 64)
	r.H, _ = strconv.ParseFloat(vals[3], 64)

	return r
}

// Valid checks for (a) a known region type, and then (b) verifies that the
// values are valid for the given type.  There is no attempt to check for
// per-image correctness, just general validity.
func (r Region) Valid() bool {
	switch r.Type {
	case RTNone:
		return false
	case RTFull, RTSquare:
		return true
	}

	if r.W <= 0 || r.H <= 0 || r.X < 0 || r.Y < 0 {
		return false
	}

	return true
}
