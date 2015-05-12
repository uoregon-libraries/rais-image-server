package iiif

import (
	"github.com/uoregon-libraries/newspaper-jp2-viewer/color-assert"
	"testing"
)

func TestSimpleInfoProfile(t *testing.T) {
	fs := FeatureSet1()
	i := fs.Info()
	assert.Equal(1, len(i.Profile), "Profile has one field", t)
	assert.Equal("http://iiif.io/api/image/2/level1.json", i.Profile[0], "Profile is level 1", t)
}
