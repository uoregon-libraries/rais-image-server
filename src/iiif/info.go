package iiif

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ProfileWrapper is a structure which has to custom-marshal itself to provide
// the rather awful profile IIIF 2.1 defines: it's an array with any number of
// elements, where the first element is always a string (URI to the conformance
// level) and subsequent elements are complex structures defining further
// capabilities.  For IIIF 3.0 the profile is a single string (the conformance
// level), and the extra capabilities live in top-level extra* properties; the
// same data is held here and serialized differently by Info.
type ProfileWrapper struct {
	profileElement2
	ConformanceURL string
}

// profileElement2 holds the pieces of the profile which can be marshaled into
// JSON without crazy pain - they just have to be marshaled as the second
// element in the aforementioned typeless profile array (v2), or split out into
// top-level / extra* properties (v3).
type profileElement2 struct {
	Formats   []string `json:"formats,omitempty"`
	Qualities []string `json:"qualities,omitempty"`
	Supports  []string `json:"supports,omitempty"`
	MaxArea   int64    `json:"maxArea,omitempty"`
	MaxWidth  int      `json:"maxWidth,omitempty"`
	MaxHeight int      `json:"maxHeight,omitempty"`
}

// MarshalJSON implements json.Marshaler
func (p *ProfileWrapper) MarshalJSON() ([]byte, error) {
	var hack = make([]any, 2)
	hack[0] = p.ConformanceURL
	hack[1] = p.profileElement2

	return json.Marshal(hack)
}

// UnmarshalJSON implements json.Unmarshaler
func (p *ProfileWrapper) UnmarshalJSON(data []byte) error {
	var hack = make([]any, 2)
	hack[0] = ""
	hack[1] = &profileElement2{}

	var err = json.Unmarshal(data, &hack)
	if err != nil {
		return err
	}

	switch v := hack[0].(type) {
	case string:
		p.ConformanceURL = v
	default:
		return fmt.Errorf("profile[0] (%#v) should have been a string", v)
	}

	switch v := hack[1].(type) {
	case *profileElement2:
		p.profileElement2 = *v
	default:
		return fmt.Errorf("profile[1] (%#v) should have been a structure", v)
	}

	return nil
}

// Info represents the simplest possible data to provide a valid IIIF
// information JSON response.  It holds the data in a version-neutral form;
// MarshalJSON and UnmarshalJSON emit/read the v2 or v3 shape based on version.
type Info struct {
	version  Version
	Context  string
	ID       string
	Type     string
	Protocol string
	Width    int
	Height   int
	Tiles    []TileSize
	Profile  ProfileWrapper
}

// NewInfo returns the static *Info data that's the same for any info response
// of the given spec version
func NewInfo(v Version) *Info {
	i := &Info{
		version:  v,
		Context:  v.ContextURI(),
		Protocol: "http://iiif.io/api/image",
	}
	if v == V3 {
		i.Type = "ImageService3"
	}
	return i
}

// Version returns the spec version this Info will serialize as
func (i *Info) Version() Version {
	return i.version
}

// infoV2 is the on-the-wire shape of a IIIF 2.1 info.json document
type infoV2 struct {
	Context  string         `json:"@context"`
	ID       string         `json:"@id"`
	Protocol string         `json:"protocol"`
	Width    int            `json:"width"`
	Height   int            `json:"height"`
	Tiles    []TileSize     `json:"tiles,omitempty"`
	Profile  ProfileWrapper `json:"profile"`
}

// infoV3 is the on-the-wire shape of a IIIF 3.0 info.json document
type infoV3 struct {
	Context        string     `json:"@context"`
	ID             string     `json:"id"`
	Type           string     `json:"type"`
	Protocol       string     `json:"protocol"`
	Profile        string     `json:"profile"`
	Width          int        `json:"width"`
	Height         int        `json:"height"`
	MaxArea        int64      `json:"maxArea,omitempty"`
	MaxWidth       int        `json:"maxWidth,omitempty"`
	MaxHeight      int        `json:"maxHeight,omitempty"`
	Tiles          []TileSize `json:"tiles,omitempty"`
	ExtraFormats   []string   `json:"extraFormats,omitempty"`
	ExtraQualities []string   `json:"extraQualities,omitempty"`
	ExtraFeatures  []string   `json:"extraFeatures,omitempty"`
}

// MarshalJSON implements json.Marshaler, emitting the v2 or v3 document shape.
// Note the on-the-wire structs are marshaled by pointer so that the addressable
// ProfileWrapper field invokes its (pointer-receiver) MarshalJSON.
func (i *Info) MarshalJSON() ([]byte, error) {
	if i.version == V3 {
		return json.Marshal(&infoV3{
			Context:        i.Context,
			ID:             i.ID,
			Type:           i.Type,
			Protocol:       i.Protocol,
			Profile:        i.Profile.ConformanceURL,
			Width:          i.Width,
			Height:         i.Height,
			MaxArea:        i.Profile.MaxArea,
			MaxWidth:       i.Profile.MaxWidth,
			MaxHeight:      i.Profile.MaxHeight,
			Tiles:          i.Tiles,
			ExtraFormats:   i.Profile.Formats,
			ExtraQualities: i.Profile.Qualities,
			ExtraFeatures:  i.Profile.Supports,
		})
	}

	return json.Marshal(&infoV2{
		Context:  i.Context,
		ID:       i.ID,
		Protocol: i.Protocol,
		Width:    i.Width,
		Height:   i.Height,
		Tiles:    i.Tiles,
		Profile:  i.Profile,
	})
}

// UnmarshalJSON implements json.Unmarshaler.  It reads the v2 or v3 shape based
// on the Info's version, which the caller must set before unmarshaling (e.g.,
// via NewInfo).  This is only exercised by the "-info.json" override path.
func (i *Info) UnmarshalJSON(data []byte) error {
	if i.version == V3 {
		var raw infoV3
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		i.Context = raw.Context
		i.ID = raw.ID
		i.Type = raw.Type
		i.Protocol = raw.Protocol
		i.Width = raw.Width
		i.Height = raw.Height
		i.Tiles = raw.Tiles
		i.Profile = ProfileWrapper{ConformanceURL: raw.Profile}
		i.Profile.MaxArea = raw.MaxArea
		i.Profile.MaxWidth = raw.MaxWidth
		i.Profile.MaxHeight = raw.MaxHeight
		i.Profile.Formats = raw.ExtraFormats
		i.Profile.Qualities = raw.ExtraQualities
		i.Profile.Supports = raw.ExtraFeatures
		return nil
	}

	var raw infoV2
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	i.Context = raw.Context
	i.ID = raw.ID
	i.Protocol = raw.Protocol
	i.Width = raw.Width
	i.Height = raw.Height
	i.Tiles = raw.Tiles
	i.Profile = raw.Profile
	return nil
}

// Info returns the default structure for a FeatureSet's info response JSON for
// the given spec version.  The caller is responsible for filling in
// image-specific values (ID and dimensions).
func (fs *FeatureSet) Info(v Version) *Info {
	i := NewInfo(v)
	i.Profile = fs.Profile(v)

	return i
}

// baseFeatureSet returns the largest standard FeatureSet that fs fully includes
// for the given spec version, along with that level's numeric value (0, 1, or
// 2) used to build the conformance label.
func (fs *FeatureSet) baseFeatureSet(v Version) (*FeatureSet, int) {
	var level2, level1, level0 *FeatureSet
	if v == V3 {
		level2, level1, level0 = FeatureSet2V3(), FeatureSet1V3(), FeatureSet0V3()
	} else {
		level2, level1, level0 = FeatureSet2(), FeatureSet1(), FeatureSet0()
	}

	if fs.includesVersion(level2, v) {
		return level2, 2
	}
	if fs.includesVersion(level1, v) {
		return level1, 1
	}
	return level0, 0
}

// Profile examines the features in the FeatureSet to determine first which
// level the FeatureSet supports, then adds any variances, formatted for the
// given spec version.
func (fs *FeatureSet) Profile(v Version) ProfileWrapper {
	baseFS, level := fs.baseFeatureSet(v)
	p := ProfileWrapper{ConformanceURL: v.conformanceLabel(level)}

	_, extraFeatures, _ := featureCompare(fs, baseFS, v)
	if len(extraFeatures) > 0 {
		p.profileElement2 = extraProfileFromFeaturesMap(extraFeatures)
	}

	return p
}

func extraProfileFromFeaturesMap(fm FeaturesMap) profileElement2 {
	p := profileElement2{
		Formats:   make([]string, 0),
		Qualities: make([]string, 0),
		Supports:  make([]string, 0),
	}

	// By default a FeaturesMap is created only listing enabled features, so as
	// long as that doesn't change, we can ignore the boolean
	for name := range fm {
		if Quality(name).Valid() {
			p.Qualities = append(p.Qualities, name)
			continue
		}
		if Format(name).Valid() {
			p.Formats = append(p.Formats, name)
			continue
		}

		p.Supports = append(p.Supports, name)
	}

	sort.Strings(p.Qualities)
	sort.Strings(p.Formats)
	sort.Strings(p.Supports)

	return p
}
