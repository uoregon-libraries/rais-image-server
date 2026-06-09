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

// toMapV3 converts a FeatureSet's boolean support values into a map using IIIF
// 3.0 feature names.  Several names differ from 2.1: notably the v2
// "sizeByForcedWh" (the "w,h" form) is "sizeByWh" in v3, and the v2 "sizeByWh"
// (the "!w,h" form) is "sizeByConfinedWh" in v3.  Feature names that no longer
// exist in v3 (sizeByWhListed, sizeByForcedWh, sizeAboveFull, sizeByDistortedWh)
// are intentionally omitted.  sizeUpscaling is also omitted because RAIS never
// upscales.
func (fs *FeatureSet) toMapV3() FeaturesMap {
	return FeaturesMap{
		"regionByPx":          fs.RegionByPx,
		"regionByPct":         fs.RegionByPct,
		"regionSquare":        fs.RegionSquare,
		"sizeByW":             fs.SizeByW,
		"sizeByH":             fs.SizeByH,
		"sizeByPct":           fs.SizeByPct,
		"sizeByWh":            fs.SizeByForcedWh,
		"sizeByConfinedWh":    fs.SizeByWh,
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

// featureMap returns the FeaturesMap appropriate for the given spec version
func (fs *FeatureSet) featureMap(v Version) FeaturesMap {
	if v == V3 {
		return fs.toMapV3()
	}
	return fs.toMap()
}

// FeatureCompare returns which features are in common between two FeatureSets,
// which are exclusive to a, and which are exclusive to b, using IIIF 2.1 feature
// names.  See featureCompare for details.
func FeatureCompare(a, b *FeatureSet) (union, onlyA, onlyB FeaturesMap) {
	return featureCompare(a, b, V2)
}

// featureCompare is the version-aware implementation behind FeatureCompare.  The
// returned maps will ONLY contain keys with a value of true, as opposed to the
// full list of features and true/false.  This helps to quickly determine
// equality, subset status, and superset status.
func featureCompare(a, b *FeatureSet, v Version) (union, onlyA, onlyB FeaturesMap) {
	union = make(FeaturesMap)
	onlyA = make(FeaturesMap)
	onlyB = make(FeaturesMap)

	mapA := a.featureMap(v)
	mapB := b.featureMap(v)

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

// includes returns whether or not fs includes all features in fsIncluded, using
// IIIF 2.1 feature naming
func (fs *FeatureSet) includes(fsIncluded *FeatureSet) bool {
	return fs.includesVersion(fsIncluded, V2)
}

// includesVersion returns whether or not fs includes all features in fsIncluded
// when compared under the given spec version's feature naming
func (fs *FeatureSet) includesVersion(fsIncluded *FeatureSet, v Version) bool {
	_, _, onlyYours := featureCompare(fs, fsIncluded, v)
	return len(onlyYours) == 0
}
