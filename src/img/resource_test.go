package img

import (
	"image"
	"math"
	"rais/src/iiif"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

var unlimited = Constraint{math.MaxInt32, math.MaxInt32, math.MaxInt64}

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
	var tall = &Resource{Decoder: d}
	var url, _ = iiif.NewURL("identifier/square/full/0/default.jpg")
	var _, err = tall.Apply(url, unlimited)
	assert.True(err == nil, "tall.Apply should not have errors", t)

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
	var wide = &Resource{Decoder: d}
	var url, _ = iiif.NewURL("identifier/square/full/0/default.jpg")
	var _, err = wide.Apply(url, unlimited)
	assert.True(err == nil, "wide.Apply should not have errors", t)

	assert.Equal(image.Point{650, 650}, d.crop.Size(), "square should be height x height", t)
	assert.Equal(1675, d.crop.Min.X, "wide image left", t)
	assert.Equal(0, d.crop.Min.Y, "wide image top", t)
	assert.Equal(2325, d.crop.Max.X, "wide image right", t)
	assert.Equal(650, d.crop.Max.Y, "wide image bottom", t)
}

func TestMaxSizeNoConstraints(t *testing.T) {
	var d = &fakeDecoder{w: 4000, h: 650, tw: 128, th: 128, l: 4}
	var img = &Resource{Decoder: d}
	var url, _ = iiif.NewURL("identifier/full/max/0/default.jpg")
	var _, err = img.Apply(url, unlimited)
	assert.True(err == nil, "img.Apply should not have errors", t)

	assert.Equal(image.Point{4000, 650}, d.crop.Size(), "max size should be full width x height", t)
	assert.Equal(4000, d.resizeW, "resize width", t)
	assert.Equal(650, d.resizeH, "resize height", t)
}

func TestMaxSizeConstrainWidth(t *testing.T) {
	var d = &fakeDecoder{w: 4000, h: 650, tw: 128, th: 128, l: 4}
	var img = &Resource{Decoder: d}
	var url, _ = iiif.NewURL("identifier/full/max/0/default.jpg")
	var c = unlimited
	c.Width = 400
	var _, err = img.Apply(url, c)
	assert.True(err == nil, "img.Apply should not have errors", t)

	assert.Equal(image.Point{4000, 650}, d.crop.Size(), "no crop", t)
	assert.Equal(400, d.resizeW, "resize width", t)
	assert.Equal(65, d.resizeH, "resize height", t)
}

func TestMaxSizeConstrainHeight(t *testing.T) {
	var d = &fakeDecoder{w: 4000, h: 650, tw: 128, th: 128, l: 4}
	var img = &Resource{Decoder: d}
	var url, _ = iiif.NewURL("identifier/full/max/0/default.jpg")
	var c = unlimited
	c.Height = 325
	var _, err = img.Apply(url, c)
	assert.True(err == nil, "img.Apply should not have errors", t)

	assert.Equal(image.Point{4000, 650}, d.crop.Size(), "no crop", t)
	assert.Equal(2000, d.resizeW, "resize width", t)
	assert.Equal(325, d.resizeH, "resize height", t)
}

func TestMaxSizeConstrainArea(t *testing.T) {
	var d = &fakeDecoder{w: 4000, h: 600, tw: 128, th: 128, l: 4}
	var img = &Resource{Decoder: d}
	var url, _ = iiif.NewURL("identifier/full/max/0/default.jpg")
	var c = unlimited
	c.Area = 37500
	var _, err = img.Apply(url, c)
	assert.True(err == nil, "img.Apply should not have errors", t)

	assert.Equal(image.Point{4000, 600}, d.crop.Size(), "no crop", t)
	assert.Equal(500, d.resizeW, "resize width", t)
	assert.Equal(75, d.resizeH, "resize height", t)
}
