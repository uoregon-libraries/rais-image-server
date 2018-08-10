package iiif

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ProfileWrapper is a structure which has to custom-marshal itself to provide
// the rather awful profile IIIF defines: it's an array with any number of
// elements, where the first element is always a string (URI to the conformance
// level) and subsequent elements are complex structures defining further
// capabilities.
type ProfileWrapper struct {
	profileElement2
	ConformanceURL string
}

// profileElement2 holds the pieces of the profile which can be marshaled into
// JSON without crazy pain - they just have to be marshaled as the second
// element in the aforementioned typeless profile array.
type profileElement2 struct {
	Formats   []string `json:"formats,omitempty"`
	Qualities []string `json:"qualities,omitempty"`
	Supports  []string `json:"supports,omitempty"`
}

// MarshalJSON implements json.Marshaler
func (p *ProfileWrapper) MarshalJSON() ([]byte, error) {
	var hack = make([]interface{}, 2)
	hack[0] = p.ConformanceURL
	hack[1] = p.profileElement2

	return json.Marshal(hack)
}

// UnmarshalJSON implements json.Unmarshaler
func (p *ProfileWrapper) UnmarshalJSON(data []byte) error {
	var hack = make([]interface{}, 2)
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
// information JSON response
type Info struct {
	Context  string         `json:"@context"`
	ID       string         `json:"@id"`
	Protocol string         `json:"protocol"`
	Width    int            `json:"width"`
	Height   int            `json:"height"`
	Tiles    []TileSize     `json:"tiles,omitempty"`
	Profile  ProfileWrapper `json:"profile"`
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

	return i
}

// baseFeatureSetData returns a FeatureSet instance for the base level as well
// as the profile URI for a given feature level
func (fs *FeatureSet) baseFeatureSet() (*FeatureSet, string) {
	FeaturesLevel2 := FeatureSet2()
	if fs.includes(FeaturesLevel2) {
		return FeaturesLevel2, "http://iiif.io/api/image/2/level2.json"
	}

	FeaturesLevel1 := FeatureSet1()
	if fs.includes(FeaturesLevel1) {
		return FeaturesLevel1, "http://iiif.io/api/image/2/level1.json"
	}

	return FeatureSet0(), "http://iiif.io/api/image/2/level0.json"
}

// Profile examines the features in the FeatureSet to determine first which
// level the FeatureSet supports, then adds any variances.
func (fs *FeatureSet) Profile() ProfileWrapper {
	baseFS, u := fs.baseFeatureSet()
	p := ProfileWrapper{ConformanceURL: u}

	_, extraFeatures, _ := FeatureCompare(fs, baseFS)
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
