package openjpeg

import (
	"image"
	"os"
	"testing"
)

func jp2i() *JP2Image {
	dir, _ := os.Getwd()
	jp2, err := NewJP2Image(dir + "/../test-world.jp2")
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
	if jp2.Dimensions() != image.Rect(0, 0, 800, 400) {
		t.Error("Dimensions were incorrect for the test jp2")
	}
}

func TestDirectConversion(t *testing.T) {
	jp2 := jp2i()
	i, err := jp2.DecodeImage()
	if err != nil {
		t.Errorf("jp2.DecodeImage() got an error: %#v", err)
		return
	}
	assertEqualInt(0, i.Bounds().Min.X, "Min.X should be 0", t)
	assertEqualInt(0, i.Bounds().Min.Y, "Min.Y should be 0", t)
	assertEqualInt(800, i.Bounds().Max.X, "Max.X should be 800", t)
	assertEqualInt(400, i.Bounds().Max.Y, "Max.Y should be 400", t)
}

func TestCrop(t *testing.T) {
	jp2 := jp2i()
	jp2.SetCrop(image.Rect(200, 100, 500, 400))
	i, err := jp2.DecodeImage()
	if err != nil {
		t.Errorf("jp2.DecodeImage() got an error: %#v", err)
		return
	}
	assertEqualInt(0, i.Bounds().Min.X, "Min.X should be 0", t)
	assertEqualInt(0, i.Bounds().Min.Y, "Min.Y should be 0", t)
	assertEqualInt(300, i.Bounds().Max.X, "Max.X should be 300 (cropped X from 200 - 500)", t)
	assertEqualInt(300, i.Bounds().Max.Y, "Max.Y should be 300 (cropped Y from 100 - 400)", t)
}

// This serves as a resize test as well as a test that we properly check
// maximum resolution factor
func TestResize(t *testing.T) {
	jp2 := jp2i()
	jp2.SetResize(50, 50)
	i, err := jp2.DecodeImage()
	if err != nil {
		t.Errorf("jp2.DecodeImage() got an error: %#v", err)
		return
	}

	assertEqualInt(0, i.Bounds().Min.X, "Min.X should be 0", t)
	assertEqualInt(0, i.Bounds().Min.Y, "Min.Y should be 0", t)
	assertEqualInt(50, i.Bounds().Max.X, "Max.X should be 50", t)
	assertEqualInt(50, i.Bounds().Max.Y, "Max.Y should be 50", t)
}

func TestResizeAndCrop(t *testing.T) {
	jp2 := jp2i()
	jp2.SetCrop(image.Rect(200, 100, 500, 400))
	jp2.SetResize(125, 125)
	i, err := jp2.DecodeImage()
	if err != nil {
		t.Errorf("jp2.DecodeImage() got an error: %#v", err)
		return
	}

	assertEqualInt(0, i.Bounds().Min.X, "Min.X should be 0", t)
	assertEqualInt(0, i.Bounds().Min.Y, "Min.Y should be 0", t)
	assertEqualInt(125, i.Bounds().Max.X, "Max.X should be 125", t)
	assertEqualInt(125, i.Bounds().Max.Y, "Max.Y should be 125", t)
}
