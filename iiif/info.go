package iiif

type profile interface{}

// Info represents the simplest possible data to provide a valid IIIF
// information JSON response
type Info struct {
	Context  string     `json:"@context"`
	ID       string     `json:"@id"`
	Protocol string     `json:"protocol"`
	Width    int        `json:"width"`
	Height   int        `json:"height"`
	Tiles    []TileSize `json:"tiles,omitempty"`
	Profile  []profile  `json:"profile"`
}

// NewInfo returns the static *Info data that's the same for any info response
func NewInfo() *Info {
	return &Info{
		Context:  "http://iiif.io/api/image/2/context.json",
		Protocol: "http://iiif.io/api/image",
	}
}

// Info returns the default structure for a FeatureSet's info response JSON.
// The caller is responsible for filling in image-specific values (ID and
// dimensions).
func (fs *FeatureSet) Info() *Info {
	i := NewInfo()
	i.Profile = fs.Profile()
	i.Tiles = fs.TileSizes

	return i
}

// Profile examines the features in the FeatureSet to determine first which
// level the FeatureSet supports, then adds any variances.
func (fs *FeatureSet) Profile() []profile {
	p := make([]profile, 1)

	var baseFeatureSet *FeatureSet
	if fs.includes(FeaturesLevel2) {
		baseFeatureSet = FeaturesLevel2
		p[0] = profile("http://iiif.io/api/image/2/level2.json")
	} else if fs.includes(FeaturesLevel1) {
		baseFeatureSet = FeaturesLevel1
		p[0] = profile("http://iiif.io/api/image/2/level1.json")
	} else {
		baseFeatureSet = FeaturesLevel0
		p[0] = profile("http://iiif.io/api/image/2/level0.json")
	}

	_, extraFeatures, _ := FeatureCompare(fs, baseFeatureSet)

	if len(extraFeatures) > 0 {
	}

	return p
}
