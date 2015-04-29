package main

// Possible IIIF 2.0 features.  The fields are the same as the string to report
// features, except that the first character should be lowercased.
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

	// "Quality", or as normal folk call it, "color depth"
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
}

var FeaturesLevel0 = FeatureSet{
	SizeByWhListed: true,
	Default:        true,
	Jpg:            true,
}

var FeaturesLevel1 = FeatureSet{
	RegionByPx:      true,
	SizeByWhListed:  true,
	SizeByW:         true,
	SizeByH:         true,
	SizeByPct:       true,
	Default:         true,
	Jpg:             true,
	BaseUriRedirect: true,
	Cors:            true,
	JsonldMediaType: true,
}

var FeaturesLevel2 = FeatureSet{
	RegionByPx:      true,
	RegionByPct:     true,
	SizeByWhListed:  true,
	SizeByW:         true,
	SizeByH:         true,
	SizeByPct:       true,
	SizeByForcedWh:  true,
	SizeByWh:        true,
	RotationBy90s:   true,
	Default:         true,
	Color:           true,
	Gray:            true,
	Bitonal:         true,
	Jpg:             true,
	Png:             true,
	BaseUriRedirect: true,
	Cors:            true,
	JsonldMediaType: true,
}
