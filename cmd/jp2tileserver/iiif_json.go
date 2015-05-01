package main

type tiledata struct {
	Width        int   `json:"width"`
	Height       int   `json:"height,omitempty"`
	ScaleFactors []int `json:"scaleFactors"`
}

// IIIFInfo represents the simplest possible data to provide a valid IIIF
// information JSON response
type IIIFInfo struct {
	Context  string     `json:"@context"`
	ID       string     `json:"@id"`
	Protocol string     `json:"protocol"`
	Width    int        `json:"width"`
	Height   int        `json:"height"`
	Profile  []string   `json:"profile"`
	Tiles    []tiledata `json:"tiles,omitempty"`
}

// Creates the default structure for converting to the IIIF Information JSON.
// The handler is responsible for filling in ID and dimensions.
func NewIIIFInfo() *IIIFInfo {
	i := &IIIFInfo{
		Context:  "http://iiif.io/api/image/2/context.json",
		Protocol: "http://iiif.io/api/image",
		Profile:  []string{"http://iiif.io/api/image/2/level1.json"},
		Tiles:    make([]tiledata, 0),
	}

	sf := []int{1, 2, 4, 8, 16, 32, 64}
	for _, val := range tileSizes {
		i.Tiles = append(i.Tiles, tiledata{Width: val, ScaleFactors: sf})
	}

	return i
}
