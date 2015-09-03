package iiif

import (
	"color-assert"
	"testing"
)

func TestFormatValidity(t *testing.T) {
	formats := []string{"jpg", "tif", "png", "gif", "jp2", "pdf", "webp"}
	for _, f := range formats {
		assert.True(Format(f).Valid(), f+" is a valid format", t)
	}
}
