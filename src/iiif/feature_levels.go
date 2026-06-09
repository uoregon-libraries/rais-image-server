package iiif

// FeatureSet0 returns a copy of the feature set required for a
// level-0-compliant IIIF server
func FeatureSet0() *FeatureSet {
	return &FeatureSet{
		SizeByWhListed: true,
		Default:        true,
		Jpg:            true,
	}
}

// FeatureSet1 returns a copy of the feature set required for a
// level-1-compliant IIIF server
func FeatureSet1() *FeatureSet {
	return &FeatureSet{
		RegionByPx:      true,
		SizeByWhListed:  true,
		SizeByW:         true,
		SizeByH:         true,
		SizeByPct:       true,
		Default:         true,
		Jpg:             true,
		BaseURIRedirect: true,
		Cors:            true,
		JsonldMediaType: true,
	}
}

// FeatureSet2 returns a copy of the feature set required for a
// level-2-compliant IIIF server
func FeatureSet2() *FeatureSet {
	return &FeatureSet{
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
		BaseURIRedirect: true,
		Cors:            true,
		JsonldMediaType: true,
	}
}

// FeatureSet0V3 returns a copy of the feature set required for a
// level-0-compliant IIIF 3.0 server.  Level 0 requires only the implicit
// baseline (full region, max size, no rotation, default quality, jpg format).
func FeatureSet0V3() *FeatureSet {
	return &FeatureSet{
		Default: true,
		Jpg:     true,
	}
}

// FeatureSet1V3 returns a copy of the feature set required for a
// level-1-compliant IIIF 3.0 server.  Note that regionSquare is required at
// level 1 in v3 (it was optional in v2), and that SizeByForcedWh is the internal
// boolean for the v3 "sizeByWh" (the "w,h" form).
func FeatureSet1V3() *FeatureSet {
	return &FeatureSet{
		RegionByPx:      true,
		RegionSquare:    true,
		SizeByW:         true,
		SizeByH:         true,
		SizeByForcedWh:  true,
		Default:         true,
		Jpg:             true,
		BaseURIRedirect: true,
		Cors:            true,
		JsonldMediaType: true,
	}
}

// FeatureSet2V3 returns a copy of the feature set required for a
// level-2-compliant IIIF 3.0 server.  SizeByWh is the internal boolean for the
// v3 "sizeByConfinedWh" (the "!w,h" form).
func FeatureSet2V3() *FeatureSet {
	return &FeatureSet{
		RegionByPx:      true,
		RegionByPct:     true,
		RegionSquare:    true,
		SizeByW:         true,
		SizeByH:         true,
		SizeByPct:       true,
		SizeByForcedWh:  true,
		SizeByWh:        true,
		RotationBy90s:   true,
		Default:         true,
		Color:           true,
		Gray:            true,
		Jpg:             true,
		Png:             true,
		BaseURIRedirect: true,
		Cors:            true,
		JsonldMediaType: true,
	}
}

// AllFeatures returns the complete list of everything supported by RAIS at
// this time
func AllFeatures() *FeatureSet {
	return &FeatureSet{
		RegionByPx:   true,
		RegionByPct:  true,
		RegionSquare: true,

		SizeByWhListed:    true,
		SizeByW:           true,
		SizeByH:           true,
		SizeByPct:         true,
		SizeByWh:          true,
		SizeByForcedWh:    true,
		SizeAboveFull:     true,
		SizeByConfinedWh:  true,
		SizeByDistortedWh: true,

		RotationBy90s: true,
		Mirroring:     true,

		Default: true,
		Color:   true,
		Gray:    true,
		Bitonal: true,

		Jpg: true,
		Png: true,
		Gif: false,
		Tif: true,

		BaseURIRedirect: true,
		Cors:            true,
		JsonldMediaType: true,
	}
}
