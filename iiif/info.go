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
