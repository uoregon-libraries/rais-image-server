package iiif

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

func TestRotation360(t *testing.T) {
	r := StringToRotation("360.0")
	assert.True(r.Valid(), "r.Valid", t)
	assert.Equal(0.0, r.Degrees, "r.Degrees", t)
}

func TestInvalidRotation(t *testing.T) {
	r := Rotation{Degrees: -1}
	assert.True(!r.Valid(), "!r.Valid", t)
	r = StringToRotation("!-1")
	assert.True(!r.Valid(), "!r.Valid", t)
	r = StringToRotation("360.1")
	assert.True(!r.Valid(), "!r.Valid", t)
	r = StringToRotation("!360.1")
	assert.True(!r.Valid(), "!r.Valid", t)
}
