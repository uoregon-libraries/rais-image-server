package iiif

import (
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

func TestSimpleInfoProfile(t *testing.T) {
	fs := FeatureSet1()
	i := fs.Info()
	assert.Equal("http://iiif.io/api/image/2/level1.json", i.Profile.ConformanceURL, "Profile is level 1", t)

	extra := i.Profile.profileElement2
	assert.Equal(0, len(extra.Supports), "extra supports", t)
	assert.Equal(0, len(extra.Qualities), "extra qualities", t)
	assert.Equal(0, len(extra.Formats), "extra formats", t)
}

// Removing a single item from level 1 should result in a level 0 profile that
// adds a bunch of features
func TestLevel1MissingFeaturesProfile(t *testing.T) {
	fs := FeatureSet1()
	fs.SizeByPct = false
	i := fs.Info()
	assert.Equal("http://iiif.io/api/image/2/level0.json", i.Profile.ConformanceURL, "Profile is level 0", t)

	extra := i.Profile.profileElement2
	assert.Equal(6, len(extra.Supports), "There are 6 extra features", t)
	assert.Equal(0, len(extra.Qualities), "There are 0 extra qualities", t)
	assert.Equal(0, len(extra.Formats), "There are 0 extra formats", t)
	assert.IncludesString("regionByPx", extra.Supports, "Custom FS support", t)
	assert.IncludesString("sizeByW", extra.Supports, "Custom FS support", t)
	assert.IncludesString("sizeByH", extra.Supports, "Custom FS support", t)
	assert.IncludesString("baseUriRedirect", extra.Supports, "Custom FS support", t)
	assert.IncludesString("cors", extra.Supports, "Custom FS support", t)
	assert.IncludesString("jsonldMediaType", extra.Supports, "Custom FS support", t)

	// Just for kicks, maybe let's verify some formats and qualities
	fs.Color = true
	fs.Bitonal = true
	fs.Png = true
	fs.Pdf = true
	fs.Jp2 = true
	fs.Gif = true
	i = fs.Info()
	extra = i.Profile.profileElement2
	assert.Equal(2, len(extra.Qualities), "There are 2 extra qualities now", t)
	assert.Equal(4, len(extra.Formats), "There are 4 extra formats now", t)
	assert.IncludesString("color", extra.Qualities, "Extra quality support", t)
	assert.IncludesString("bitonal", extra.Qualities, "Extra quality support", t)
	assert.IncludesString("png", extra.Formats, "Extra format support", t)
	assert.IncludesString("pdf", extra.Formats, "Extra format support", t)
	assert.IncludesString("jp2", extra.Formats, "Extra format support", t)
	assert.IncludesString("gif", extra.Formats, "Extra format support", t)
}

func TestAllFeaturesEnabled(t *testing.T) {
	fs := AllFeatures()
	i := fs.Info()
	assert.Equal("http://iiif.io/api/image/2/level2.json", i.Profile.ConformanceURL, "Profile conformance level", t)

	extra := i.Profile.profileElement2
	assert.Equal(3, len(extra.Supports), "There are 3 extra features", t)
	assert.Equal(0, len(extra.Qualities), "There are 0 extra qualities", t)
	assert.Equal(1, len(extra.Formats), "There is 1 extra format", t)
	assert.IncludesString("regionSquare", extra.Supports, "Custom FS support", t)
	assert.IncludesString("sizeAboveFull", extra.Supports, "Custom FS support", t)
	assert.IncludesString("mirroring", extra.Supports, "Custom FS support", t)
	assert.IncludesString("tif", extra.Formats, "Custom FS support", t)
}
