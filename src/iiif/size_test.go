package iiif

import (
	"image"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

func TestSizeTypeFull(t *testing.T) {
	s := StringToSize("full")
	assert.True(s.Valid(), "s.Valid()", t)
	assert.Equal(STFull, s.Type, "s.Type == STFull", t)
}

func TestSizeTypeScaleWidth(t *testing.T) {
	s := StringToSize("125,")
	assert.True(s.Valid(), "s.Valid()", t)
	assert.Equal(STScaleToWidth, s.Type, "s.Type == STScaleToWidth", t)
	assert.Equal(125, s.W, "s.W", t)
	assert.Equal(0, s.H, "s.H", t)
}

func TestSizeTypeScaleHeight(t *testing.T) {
	s := StringToSize(",250")
	assert.True(s.Valid(), "s.Valid()", t)
	assert.Equal(STScaleToHeight, s.Type, "s.Type == STScaleToHeight", t)
	assert.Equal(0, s.W, "s.W", t)
	assert.Equal(250, s.H, "s.H", t)
}

func TestSizeTypePercent(t *testing.T) {
	s := StringToSize("pct:41.6")
	assert.True(s.Valid(), "s.Valid()", t)
	assert.Equal(STScalePercent, s.Type, "s.Type == STScalePercent", t)
	assert.Equal(41.6, s.Percent, "s.Percent", t)
}

func TestSizeTypeExact(t *testing.T) {
	s := StringToSize("125,250")
	assert.True(s.Valid(), "s.Valid()", t)
	assert.Equal(STExact, s.Type, "s.Type == STExact", t)
	assert.Equal(125, s.W, "s.W", t)
	assert.Equal(250, s.H, "s.H", t)
}

func TestSizeTypeBestFit(t *testing.T) {
	s := StringToSize("!25,50")
	assert.True(s.Valid(), "s.Valid()", t)
	assert.Equal(STBestFit, s.Type, "s.Type == STBestFit", t)
	assert.Equal(25, s.W, "s.W", t)
	assert.Equal(50, s.H, "s.H", t)
}

func TestInvalidSizes(t *testing.T) {
	s := Size{}
	assert.True(!s.Valid(), "!s.Valid()", t)
	s = StringToSize(",0")
	assert.True(!s.Valid(), "!s.Valid()", t)
	s = StringToSize("0,")
	assert.True(!s.Valid(), "!s.Valid()", t)
	s = StringToSize("0,100")
	assert.True(!s.Valid(), "!s.Valid()", t)
	s = StringToSize("100,0")
	assert.True(!s.Valid(), "!s.Valid()", t)
	s = StringToSize("!0,100")
	assert.True(!s.Valid(), "!s.Valid()", t)
	s = StringToSize("!100,0")
	assert.True(!s.Valid(), "!s.Valid()", t)
	s = StringToSize("pct:0")
	assert.True(!s.Valid(), "!s.Valid()", t)
}

func TestGetResize(t *testing.T) {
	s := Size{Type: STFull}
	source := image.Rect(0, 0, 600, 1200)
	scale := s.GetResize(source)
	assert.Equal(scale.Dx(), source.Dx(), "full resize Dx", t)
	assert.Equal(scale.Dy(), source.Dy(), "full resize Dy", t)

	s.Type = STScaleToWidth
	s.W = 90
	scale = s.GetResize(source)
	assert.Equal(scale.Dx(), 90, "scale-to-width Dx", t)
	assert.Equal(scale.Dy(), 180, "scale-to-width Dy", t)

	s.Type = STScaleToHeight
	s.H = 90
	scale = s.GetResize(source)
	assert.Equal(scale.Dx(), 45, "scale-to-height Dx", t)
	assert.Equal(scale.Dy(), 90, "scale-to-height Dy", t)

	s.Type = STScalePercent
	s.Percent = 100 * 2.0 / 3.0
	scale = s.GetResize(source)
	assert.Equal(scale.Dx(), 400, "scale-to-pct Dx", t)
	assert.Equal(scale.Dy(), 800, "scale-to-pct Dy", t)

	s.Type = STExact
	s.W = 95
	s.H = 100
	scale = s.GetResize(source)
	assert.Equal(scale.Dx(), 95, "scale-to-exact Dx", t)
	assert.Equal(scale.Dy(), 100, "scale-to-exact Dy", t)

	s.Type = STBestFit
	scale = s.GetResize(source)
	assert.Equal(scale.Dx(), 50, "scale-to-pct Dx", t)
	assert.Equal(scale.Dy(), 100, "scale-to-pct Dy", t)
}
