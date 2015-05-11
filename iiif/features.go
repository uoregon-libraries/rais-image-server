package iiif

// featuresMap is a simple map for boolean features, used for comparing
// featuresets and reporting features beyond the reported level
type featuresMap map[string]bool

// TileSize represents a supported tile size for a feature set to expose.  This
// data is serialized in an info request and therefore must have JSON tags.
type TileSize struct {
	Width        int   `json:"width"`
	Height       int   `json:"height,omitempty"`
	ScaleFactors []int `json:"scaleFactors"`
}

// FeatureSet represents possible IIIF 2.0 features.  The boolean fields are
// the same as the string to report features, except that the first character
// should be lowercased.
//
// Note that using this in a different server only gets you so far.  As noted
// in the Supported() documentation below, verifying complete support is
// trickier than just checking a URL, and a server that doesn't support
// arbitrary resizing can still advertise specific sizes that will work.
type FeatureSet struct {
	// Region options: note that full isn't specified but must be supported
	RegionByPx  bool
	RegionByPct bool

	// Size options: note that full isn't specified but must be supported
	SizeByWhListed bool
	SizeByW        bool
	SizeByH        bool
	SizeByPct      bool
	SizeByForcedWh bool
	SizeByWh       bool
	SizeAboveFull  bool

	// Rotation and mirroring
	RotationBy90s     bool
	RotationArbitrary bool
	Mirroring         bool

	// "Quality" (color depth / color space)
	Default bool
	Color   bool
	Gray    bool
	Bitonal bool

	// Format
	Jpg  bool
	Png  bool
	Tif  bool
	Gif  bool
	Jp2  bool
	Pdf  bool
	Webp bool

	// HTTP features
	BaseUriRedirect     bool
	Cors                bool
	JsonldMediaType     bool
	ProfileLinkHeader   bool
	CanonicalLinkHeader bool

	// Non-boolean feature support
	TileSizes []TileSize
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

// toMap converts a FeatureSet's boolean support values into a map suitable for
// use in comparison to other feature sets.  The strings used are lowercased so
// they can be used as-is within "formats", "qualities", and/or "supports"
// arrays.
func (fs *FeatureSet) toMap() featuresMap {
	return featuresMap{
		"regionByPx":          fs.RegionByPx,
		"regionByPct":         fs.RegionByPct,
		"sizeByWhListed":      fs.SizeByWhListed,
		"sizeByW":             fs.SizeByW,
		"sizeByH":             fs.SizeByH,
		"sizeByPct":           fs.SizeByPct,
		"sizeByForcedWh":      fs.SizeByForcedWh,
		"sizeByWh":            fs.SizeByWh,
		"sizeAboveFull":       fs.SizeAboveFull,
		"rotationBy90s":       fs.RotationBy90s,
		"rotationArbitrary":   fs.RotationArbitrary,
		"mirroring":           fs.Mirroring,
		"default":             fs.Default,
		"color":               fs.Color,
		"gray":                fs.Gray,
		"bitonal":             fs.Bitonal,
		"jpg":                 fs.Jpg,
		"png":                 fs.Png,
		"tif":                 fs.Tif,
		"gif":                 fs.Gif,
		"jp2":                 fs.Jp2,
		"pdf":                 fs.Pdf,
		"webp":                fs.Webp,
		"baseUriRedirect":     fs.BaseUriRedirect,
		"cors":                fs.Cors,
		"jsonldMediaType":     fs.JsonldMediaType,
		"profileLinkHeader":   fs.ProfileLinkHeader,
		"canonicalLinkHeader": fs.CanonicalLinkHeader,
	}
}

// FeatureCompare returns which features are in common between two FeatureSets,
// which are exclusive to a, and which are exclusive to b.  The returned maps
// will ONLY contain keys with a value of true, as opposed to the full list of
// features and true/false.  This helps to quickly determine equality, subset
// status, and superset status.
func FeatureCompare(a, b *FeatureSet) (union, onlyA, onlyB featuresMap) {
	union = make(featuresMap)
	onlyA = make(featuresMap)
	onlyB = make(featuresMap)

	mapA := a.toMap()
	mapB := b.toMap()

	for feature, supportedA := range mapA {
		supportedB := mapB[feature]
		if supportedA && supportedB {
			union[feature] = true
			continue
		}

		if supportedA {
			onlyA[feature] = true
			continue
		}

		if supportedB {
			onlyB[feature] = true
		}
	}

	return
}

// includes returns whether or not fs includes all features in fsIncluded
func (fs *FeatureSet) includes(fsIncluded *FeatureSet) bool {
	_, _, onlyYours := FeatureCompare(fs, fsIncluded)
	return len(onlyYours) == 0
}
