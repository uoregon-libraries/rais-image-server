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

// StringToRotation creates a Rotation from a string as seen in an IIIF URL.
// An invalid string would result in a 0-degree rotation as opposed to an error
// condition.  This is a known issue which needs to be fixed.
func StringToRotation(p string) Rotation {
	r := Rotation{}
	if p[0:1] == "!" {
		r.Mirror = true
		p = p[1:]
	}

	r.Degrees, _ = strconv.ParseFloat(p, 64)

	// This isn't actually to spec, but it makes way more sense than only
	// allowing 360 for compliance level 2 (and in fact *requiring* it there)
	if r.Degrees == 360 {
		r.Degrees = 0
	}

	return r
}

// Valid just returns whether or not the degrees value is within a sane range:
// 0 <= r.Degrees < 360
func (r Rotation) Valid() bool {
	return r.Degrees >= 0 && r.Degrees < 360
}
