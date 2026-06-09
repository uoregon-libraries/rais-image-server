package iiif

import (
	"encoding/json"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

// TestSizeVersionValidity verifies the version-specific legality of the size
// parameter: v2 allows "full" but not the "^" upscaling prefix, while v3 allows
// "max" (and "^") but not "full".
func TestSizeVersionValidity(t *testing.T) {
	var id = "identifier"

	// "full" is fine in v2 but invalid in v3
	u2, err := NewURL(id+"/full/full/0/default.jpg", V2)
	assert.NilError(err, "v2 accepts size=full", t)
	assert.Equal(STFull, u2.Size.Type, "v2 size is STFull", t)

	_, err = NewURL(id+"/full/full/0/default.jpg", V3)
	assert.True(err != nil, "v3 rejects size=full", t)

	// "max" is valid in both versions
	u3, err := NewURL(id+"/full/max/0/default.jpg", V3)
	assert.NilError(err, "v3 accepts size=max", t)
	assert.Equal(STMax, u3.Size.Type, "v3 size is STMax", t)

	// The "^" upscaling prefix is valid syntax only in v3
	uUp, err := NewURL(id+"/full/^max/0/default.jpg", V3)
	assert.NilError(err, "v3 accepts ^ upscaling prefix syntactically", t)
	assert.True(uUp.Size.Upscale, "v3 ^ sets Upscale", t)
	assert.Equal(STMax, uUp.Size.Type, "v3 ^max is still STMax", t)

	_, err = NewURL(id+"/full/^max/0/default.jpg", V2)
	assert.True(err != nil, "v2 rejects ^ upscaling prefix", t)
}

// TestStringToSizeUpscale verifies the "^" prefix is parsed for any size form
func TestStringToSizeUpscale(t *testing.T) {
	s := StringToSize("^500,")
	assert.True(s.Upscale, "^500, sets Upscale", t)
	assert.Equal(STScaleToWidth, s.Type, "^500, is scale-to-width", t)
	assert.Equal(500, s.W, "^500, width", t)

	s = StringToSize("500,")
	assert.True(!s.Upscale, "500, does not set Upscale", t)
}

// TestInfoV3Shape verifies the IIIF 3.0 info.json document shape
func TestInfoV3Shape(t *testing.T) {
	fs := AllFeatures()
	i := fs.Info(V3)
	i.ID = "https://example.com/iiif/v3/some-id"
	i.Width = 6000
	i.Height = 4000

	data, err := json.Marshal(i)
	assert.NilError(err, "marshal v3 info", t)

	var raw map[string]any
	assert.NilError(json.Unmarshal(data, &raw), "unmarshal v3 info into a map", t)

	assert.Equal("http://iiif.io/api/image/3/context.json", raw["@context"], "v3 context", t)
	assert.Equal("ImageService3", raw["type"], "v3 type", t)
	assert.Equal("https://example.com/iiif/v3/some-id", raw["id"], "v3 id (not @id)", t)
	assert.Equal("level2", raw["profile"], "v3 profile is a level string", t)
	assert.Equal(nil, raw["@id"], "v3 must not emit @id", t)

	_, hasExtraFeatures := raw["extraFeatures"]
	assert.True(hasExtraFeatures, "v3 reports extraFeatures", t)
}

// TestInfoV3ExtraFeatureNames verifies the v3 feature-name remapping for the
// size forms that were renamed between v2 and v3
func TestInfoV3ExtraFeatureNames(t *testing.T) {
	// A level-2 v3 server with mirroring is the simplest way to get a non-empty
	// extraFeatures list using a renamed-but-required feature name elsewhere.
	fs := FeatureSet2V3()
	fs.Mirroring = true
	p := fs.Profile(V3)

	assert.IncludesString("mirroring", p.Supports, "mirroring is an extra feature", t)

	// And confirm the renamed size features land under their v3 names when they
	// are extras (here using a level-1 base so the level-2 size adds show up).
	fs1 := FeatureSet1V3()
	fs1.SizeByWh = true // internal bool for v3 "sizeByConfinedWh"
	p1 := fs1.Profile(V3)
	assert.IncludesString("sizeByConfinedWh", p1.Supports, "v3 !w,h is sizeByConfinedWh", t)
}
