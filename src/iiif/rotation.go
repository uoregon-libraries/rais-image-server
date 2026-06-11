package iiif

import (
	"strconv"
)

// Rotation represents the degrees of rotation and whether or not an image is
// mirrored, as both are defined in IIIF 2.0 as being part of the rotation
// parameter in IIIF URL requests.
type Rotation struct {
	Mirror  bool
	Degrees float64
}

// StringToRotation creates a Rotation from a string as seen in a IIIF URL. A
// string which can't be parsed as a number (after the optional leading "!")
// results in a Rotation which reports itself as invalid.
func StringToRotation(p string) Rotation {
	r := Rotation{}
	if p == "" {
		r.Degrees = -1
		return r
	}
	if p[0:1] == "!" {
		r.Mirror = true
		p = p[1:]
	}

	var err error
	r.Degrees, err = strconv.ParseFloat(p, 64)
	if err != nil {
		r.Degrees = -1
		return r
	}

	// This isn't actually to spec, but it makes way more sense than only
	// allowing 360 for compliance level 2 (and in fact *requiring* it there)
	if r.Degrees == 360 {
		r.Degrees = 0
	}

	return r
}

// Valid returns false if the rotation string couldn't be parsed as a number,
// or the degrees value is outside the sane range: 0 <= r.Degrees < 360
func (r Rotation) Valid() bool {
	return r.Degrees >= 0 && r.Degrees < 360
}
