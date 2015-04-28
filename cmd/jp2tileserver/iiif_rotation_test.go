package main

import (
	"github.com/uoregon-libraries/newspaper-jp2-viewer/color-assert"
	"testing"
)

func TestRotationNormal(t *testing.T) {
	r := StringToRotation("250.5")
	assert.Equal(250.5, r.Degrees, "r.Degrees", t)
	assert.True(!r.Mirror, "!r.Mirror", t)
}

func TestRotationMirrored(t *testing.T) {
	r := StringToRotation("!90")
	assert.Equal(90.0, r.Degrees, "r.Degrees", t)
	assert.True(r.Mirror, "r.Mirror", t)
}
