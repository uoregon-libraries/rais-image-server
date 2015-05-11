package iiif

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
