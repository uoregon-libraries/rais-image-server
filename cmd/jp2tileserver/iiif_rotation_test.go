package main

import (
	"testing"
)

func TestRotationNormal(t *testing.T) {
	r := StringToRotation("250.5")
	assertEqual(250.5, r.Degrees, "r.Degrees", t)
	assert(!r.Mirror, "!r.Mirror", t)
}

func TestRotationMirrored(t *testing.T) {
	r := StringToRotation("!90")
	assertEqual(90.0, r.Degrees, "r.Degrees", t)
	assert(r.Mirror, "r.Mirror", t)
}
