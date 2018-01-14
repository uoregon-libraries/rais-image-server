package openjpeg

import (
	"color-assert"
	"image"
	"os"
	"testing"

	"github.com/uoregon-libraries/gopkg/logger"
)

func init() {
	Logger = logger.New(logger.Warn)
}

func jp2i() *JP2Image {
	dir, _ := os.Getwd()
	jp2, err := NewJP2Image(dir + "/../../docker/images/testfile/test-world.jp2")
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

// BenchmarkReadAndDecodeImage does a benchmark against every step of the
// process to simulate the parts of a tile request controlled by the openjpeg
// package: loading the JP2, setting the crop and resize, and decoding to a raw
// image resource.  The test image has no tiling, so this is benchmarking the
// most expensive operation we currently have.
func BenchmarkReadAndDecodeImage(b *testing.B) {
	for n := 0; n < b.N; n++ {
		startx := n % 512
		endx := startx + 256
		jp2 := jp2i()
		jp2.SetCrop(image.Rect(startx, 0, endx, 256))
		jp2.SetResizeWH(128, 128)
		_, err := jp2.DecodeImage()
		if err != nil {
			panic(err)
		}
	}
}

// BenchmarkReadAndDecodeImage does a benchmark against a large, tiled image to
// see how we perform when using the best-case image type
func BenchmarkReadAndDecodeTiledImage(b *testing.B) {
	dir, _ := os.Getwd()
	bigImage := dir + "/../../docker/images/jp2tests/sn00063609-19091231.jp2"

	for n := 0; n < b.N; n++ {
		var size = ((n%2)+1) * 1024
		startTileX := n % 2
		startTileY := (n / 2) % 3
		startX := startTileX * size
		endX := startX + size
		startY := startTileY * size
		endY := startY + size

	  jp2, err := NewJP2Image(bigImage)
		if err != nil {
			panic(err)
		}
		jp2.SetCrop(image.Rect(startX, startY, endX, endY))
		jp2.SetResizeWH(1024, 1024)
		if _, err := jp2.DecodeImage(); err != nil {
			panic(err)
		}
	}
}
