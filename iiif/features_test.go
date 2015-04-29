package iiif

import (
	"github.com/uoregon-libraries/newspaper-jp2-viewer/color-assert"
	"testing"
)

func TestRegionSupport(t *testing.T) {
	r := Region{Type: RTFull}
	assert.True(FeaturesLevel0.SupportsRegion(r), "RTFull supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsRegion(r), "RTFull supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsRegion(r), "RTFull supported by FL2", t)

	r.Type = RTPixel
	assert.False(FeaturesLevel0.SupportsRegion(r), "RTPixel NOT supported by FL0", t)
	assert.True(FeaturesLevel1.SupportsRegion(r), "RTPixel supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsRegion(r), "RTPixel supported by FL2", t)

	r.Type = RTPercent
	assert.False(FeaturesLevel0.SupportsRegion(r), "RTPercent NOT supported by FL0", t)
	assert.False(FeaturesLevel1.SupportsRegion(r), "RTPercent NOT supported by FL1", t)
	assert.True(FeaturesLevel2.SupportsRegion(r), "RTPercent supported by FL2", t)
}
