package iiif

// FeaturesLevel0: the required features for a level-0-compliant IIIF server
var FeaturesLevel0 = &FeatureSet{
	SizeByWhListed: true,
	Default:        true,
	Jpg:            true,
}

// FeaturesLevel1: the required features for a level-1-compliant IIIF server
var FeaturesLevel1 = &FeatureSet{
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

// FeaturesLevel2: the required features for a level-2-compliant IIIF server
var FeaturesLevel2 = &FeatureSet{
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
