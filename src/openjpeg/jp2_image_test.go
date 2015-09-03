package openjpeg

import (
	"color-assert"
	"image"
	"os"
	"testing"
)

func jp2i() *JP2Image {
	dir, _ := os.Getwd()
	jp2, err := NewJP2Image(dir + "/../testfile/test-world.jp2")
	if err != nil {
		panic("Error reading JP2 for testing!")
	}
	return jp2
}

func TestNewJP2Image(t *testing.T) {
	jp2 := jp2i()

	if jp2 == nil {
		t.Error("No JP2 object!")
	}

	t.Log(jp2.image)
}

func TestDimensions(t *testing.T) {
	jp2 := jp2i()
	jp2.ReadHeader()
	assert.Equal(800, jp2.GetWidth(), "jp2 width is 800px", t)
	assert.Equal(400, jp2.GetHeight(), "jp2 height is 400px", t)
}

func TestDirectConversion(t *testing.T) {
	jp2 := jp2i()
	i, err := jp2.DecodeImage()
	assert.Equal(err, nil, "No error decoding jp2", t)
	assert.Equal(0, i.Bounds().Min.X, "Min.X should be 0", t)
	assert.Equal(0, i.Bounds().Min.Y, "Min.Y should be 0", t)
	assert.Equal(800, i.Bounds().Max.X, "Max.X should be 800", t)
	assert.Equal(400, i.Bounds().Max.Y, "Max.Y should be 400", t)
}

func TestCrop(t *testing.T) {
	jp2 := jp2i()
	jp2.SetCrop(image.Rect(200, 100, 500, 400))
	i, err := jp2.DecodeImage()
	assert.Equal(err, nil, "No error decoding jp2", t)
	assert.Equal(0, i.Bounds().Min.X, "Min.X should be 0", t)
	assert.Equal(0, i.Bounds().Min.Y, "Min.Y should be 0", t)
	assert.Equal(300, i.Bounds().Max.X, "Max.X should be 300 (cropped X from 200 - 500)", t)
	assert.Equal(300, i.Bounds().Max.Y, "Max.Y should be 300 (cropped Y from 100 - 400)", t)
}

// This serves as a resize test as well as a test that we properly check
// maximum resolution factor
func TestResizeWH(t *testing.T) {
	jp2 := jp2i()
	jp2.SetResizeWH(50, 50)
	i, err := jp2.DecodeImage()
	assert.Equal(err, nil, "No error decoding jp2", t)
	assert.Equal(0, i.Bounds().Min.X, "Min.X should be 0", t)
	assert.Equal(0, i.Bounds().Min.Y, "Min.Y should be 0", t)
	assert.Equal(50, i.Bounds().Max.X, "Max.X should be 50", t)
	assert.Equal(50, i.Bounds().Max.Y, "Max.Y should be 50", t)
}

func TestResizeWHAndCrop(t *testing.T) {
	jp2 := jp2i()
	jp2.SetCrop(image.Rect(200, 100, 500, 400))
	jp2.SetResizeWH(125, 125)
	i, err := jp2.DecodeImage()
	assert.Equal(err, nil, "No error decoding jp2", t)
	assert.Equal(0, i.Bounds().Min.X, "Min.X should be 0", t)
	assert.Equal(0, i.Bounds().Min.Y, "Min.Y should be 0", t)
	assert.Equal(125, i.Bounds().Max.X, "Max.X should be 125", t)
	assert.Equal(125, i.Bounds().Max.Y, "Max.Y should be 125", t)
}
