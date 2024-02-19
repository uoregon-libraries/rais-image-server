package iiif

// FeaturesMap is a simple map for boolean features, used for comparing
// featuresets and reporting features beyond the reported level
type FeaturesMap map[string]bool

// TileSize represents a supported tile size for a feature set to expose.  This
// data is serialized in an info request and therefore must have JSON tags.
type TileSize struct {
	Width        int   `json:"width"`
	Height       int   `json:"height,omitempty"`
	ScaleFactors []int `json:"scaleFactors"`
}

// FeatureSet represents possible IIIF 2.1 features.  The boolean fields are
// the same as the string to report features, except that the first character
// should be lowercased.
//
// Note that using this in a different server only gets you so far.  As noted
// in the Supported() documentation below, verifying complete support is
// trickier than just checking a URL, and a server that doesn't support
// arbitrary resizing can still advertise specific sizes that will work.
type FeatureSet struct {
	// Region options: note that full isn't specified but must be supported
	RegionByPx   bool
	RegionByPct  bool
	RegionSquare bool

	// Size options: note that full isn't specified but must be supported
	SizeByWhListed    bool
	SizeByW           bool
	SizeByH           bool
	SizeByPct         bool
	SizeByForcedWh    bool
	SizeByWh          bool
	SizeAboveFull     bool
	SizeByConfinedWh  bool
	SizeByDistortedWh bool

	// Rotation and mirroring
	RotationBy90s     bool
	RotationArbitrary bool
	Mirroring         bool

	// "Quality" (color model / color depth)
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
	BaseURIRedirect     bool
	Cors                bool
	JsonldMediaType     bool
	ProfileLinkHeader   bool
	CanonicalLinkHeader bool

	// Non-boolean feature support
	TileSizes []TileSize
}

// toMap converts a FeatureSet's boolean support values into a map suitable for
// use in comparison to other feature sets.  The strings used are lowercased so
// they can be used as-is within "formats", "qualities", and/or "supports"
// arrays.
func (fs *FeatureSet) toMap() FeaturesMap {
	return FeaturesMap{
		"regionByPx":          fs.RegionByPx,
		"regionByPct":         fs.RegionByPct,
		"regionSquare":        fs.RegionSquare,
		"sizeByWhListed":      fs.SizeByWhListed,
		"sizeByW":             fs.SizeByW,
		"sizeByH":             fs.SizeByH,
		"sizeByPct":           fs.SizeByPct,
		"sizeByForcedWh":      fs.SizeByForcedWh,
		"sizeByWh":            fs.SizeByWh,
		"sizeByConfinedWh":    fs.SizeByConfinedWh,
		"sizeByDistortedWh":   fs.SizeByDistortedWh,
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
		"baseUriRedirect":     fs.BaseURIRedirect,
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
func FeatureCompare(a, b *FeatureSet) (union, onlyA, onlyB FeaturesMap) {
	union = make(FeaturesMap)
	onlyA = make(FeaturesMap)
	onlyB = make(FeaturesMap)

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

	return union, onlyA, onlyB
}

// includes returns whether or not fs includes all features in fsIncluded
func (fs *FeatureSet) includes(fsIncluded *FeatureSet) bool {
	_, _, onlyYours := FeatureCompare(fs, fsIncluded)
	return len(onlyYours) == 0
}
