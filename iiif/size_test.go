package iiif

import (
	"github.com/uoregon-libraries/newspaper-jp2-viewer/color-assert"
	"testing"
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
