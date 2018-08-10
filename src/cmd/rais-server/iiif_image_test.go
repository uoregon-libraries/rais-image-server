package main

import (
	"image"
	"rais/src/iiif"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

type fakeDecoder struct {
	// Fake image dimensions and other metadata
	w, h   int
	tw, th int
	l      int

	// Settings touched by Apply
	crop    image.Rectangle
	resizeW int
	resizeH int
}

func (d *fakeDecoder) DecodeImage() (image.Image, error) { return nil, nil }
func (d *fakeDecoder) GetWidth() int                     { return d.w }
func (d *fakeDecoder) GetHeight() int                    { return d.h }
func (d *fakeDecoder) GetTileWidth() int                 { return d.tw }
func (d *fakeDecoder) GetTileHeight() int                { return d.th }
func (d *fakeDecoder) GetLevels() int                    { return d.l }
func (d *fakeDecoder) SetCrop(rect image.Rectangle)      { d.crop = rect }
func (d *fakeDecoder) SetResizeWH(w, h int)              { d.resizeW, d.resizeH = w, h }

func TestSquareRegionTall(t *testing.T) {
	var d = &fakeDecoder{w: 400, h: 950, tw: 64, th: 64, l: 1}
	var tall = &ImageResource{Decoder: d}
	var url = iiif.NewURL("/iiif/identifier/square/full/0/default.jpg")
	var _, err = tall.Apply(url)
	assert.NilError(err, "tall.Apply should not have errors", t)

	assert.Equal(image.Point{400, 400}, d.crop.Size(), "square should be width x width", t)
	assert.Equal(0, d.crop.Min.X, "tall image left", t)
	assert.Equal(275, d.crop.Min.Y, "tall image top", t)
	assert.Equal(400, d.crop.Max.X, "tall image right", t)
	assert.Equal(675, d.crop.Max.Y, "tall image bottom", t)
}

// Now repeat it all but with a wide image; other changes just prove tile
// sizes and levels don't matter here
func TestSquareRegionWide(t *testing.T) {
	var d = &fakeDecoder{w: 4000, h: 650, tw: 128, th: 128, l: 4}
	var wide = &ImageResource{Decoder: d}
	var url = iiif.NewURL("/iiif/identifier/square/full/0/default.jpg")
	var _, err = wide.Apply(url)
	assert.NilError(err, "wide.Apply should not have errors", t)

	assert.Equal(image.Point{650, 650}, d.crop.Size(), "square should be height x height", t)
	assert.Equal(1675, d.crop.Min.X, "wide image left", t)
	assert.Equal(0, d.crop.Min.Y, "wide image top", t)
	assert.Equal(2325, d.crop.Max.X, "wide image right", t)
	assert.Equal(650, d.crop.Max.Y, "wide image bottom", t)
}
