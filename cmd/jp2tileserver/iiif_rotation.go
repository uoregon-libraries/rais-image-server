package main

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

	return r
}

