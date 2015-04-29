package iiif

import (
	"strconv"
)

type Rotation struct {
	Mirror  bool
	Degrees float64
}

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

func (r Rotation) Valid() bool {
	return r.Degrees >= 0 && r.Degrees < 360
}
