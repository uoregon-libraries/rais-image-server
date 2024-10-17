package openjpeg

// #cgo pkg-config: libopenjp2
// #include <openjpeg.h>
import "C"

import (
	"fmt"
	"image"
	"rais/src/img"
	"rais/src/jp2info"
	"reflect"
	"unsafe"

	"github.com/nfnt/resize"
)

// JP2Image is a container for our simple JP2 operations
type JP2Image struct {
	id           uint64
	streamer     img.Streamer
	info         *jp2info.Info
	decodeWidth  int
	decodeHeight int
	decodeArea   image.Rectangle
	srcRect      image.Rectangle
}

// NewJP2Image reads basic information about a file and returns a decode-ready
// JP2Image instance
func NewJP2Image(s img.Streamer) (*JP2Image, error) {
	s.Seek(0, 0)
	var info, err = new(jp2info.Scanner).ScanStream(s)
	if err != nil {
		s.Close()
		return nil, err
	}

	var i = &JP2Image{streamer: s, info: info}
	storeImage(i)

	return i, err
}

// SetResizeWH sets the image to scale to the given width and height.  If one
// dimension is 0, the decoded image will preserve the aspect ratio while
// scaling to the non-zero dimension.
func (i *JP2Image) SetResizeWH(width, height int) {
	i.decodeWidth = width
	i.decodeHeight = height
}

// SetCrop sets the image crop area for decoding an image
func (i *JP2Image) SetCrop(r image.Rectangle) {
	i.decodeArea = r
}

// DecodeImage returns an image.Image that holds the decoded image data,
// resized and cropped if resizing or cropping was requested.  Both cropping
// and resizing happen here due to the nature of openjpeg, so SetScale,
// SetResizeWH, and SetCrop must be called before this function.
func (i *JP2Image) DecodeImage() (im image.Image, err error) {
	i.computeDecodeParameters()

	var jp2 *C.opj_image_t
	jp2, err = i.rawDecode()
	// We have to clean up the jp2 memory even if we had an error due to how the
	// openjpeg APIs work
	defer C.opj_image_destroy(jp2)
	if err != nil {
		return nil, err
	}

	var comps []C.opj_image_comp_t
	var compsSlice = (*reflect.SliceHeader)((unsafe.Pointer(&comps)))
	compsSlice.Cap = int(jp2.numcomps)
	compsSlice.Len = int(jp2.numcomps)
	compsSlice.Data = uintptr(unsafe.Pointer(jp2.comps))

	var j *opjp2
	j, err = newOpjp2(comps, i.info.BPC)
	if err != nil {
		return nil, err
	}
	var i2 image.Image
	i2, err= j.decode()
	if err != nil {
		return nil, fmt.Errorf("decoding raw JP2 data: %w", err)
	}

	if i.decodeWidth != i.decodeArea.Dx() || i.decodeHeight != i.decodeArea.Dy() {
		i2 = resize.Resize(uint(i.decodeWidth), uint(i.decodeHeight), i2, resize.Bilinear)
	}
	return i2, nil
}

// GetWidth returns the image width
func (i *JP2Image) GetWidth() int {
	return int(i.info.Width)
}

// GetHeight returns the image height
func (i *JP2Image) GetHeight() int {
	return int(i.info.Height)
}

// GetTileWidth returns the tile width
func (i *JP2Image) GetTileWidth() int {
	return int(i.info.TileWidth())
}

// GetTileHeight returns the tile height
func (i *JP2Image) GetTileHeight() int {
	return int(i.info.TileHeight())
}

// GetLevels returns the number of resolution levels
func (i *JP2Image) GetLevels() int {
	return int(i.info.Levels)
}

// computeDecodeParameters sets up decode area, decode width, and decode height
// based on the image's info
func (i *JP2Image) computeDecodeParameters() {
	if i.decodeArea == image.ZR {
		i.decodeArea = image.Rect(0, 0, int(i.info.Width), int(i.info.Height))
	}

	if i.decodeWidth == 0 && i.decodeHeight == 0 {
		i.decodeWidth = i.decodeArea.Dx()
		i.decodeHeight = i.decodeArea.Dy()
	}
}

// computeProgressionLevel gets progression level if we're resizing to specific
// dimensions (it's zero if there isn't any scaling of the output)
func (i *JP2Image) computeProgressionLevel() int {
	if i.decodeWidth == i.decodeArea.Dx() && i.decodeHeight == i.decodeArea.Dy() {
		return 0
	}

	level := desiredProgressionLevel(i.decodeArea, i.decodeWidth, i.decodeHeight)
	if level > i.GetLevels() {
		Logger.Debugf("Progression level requested (%d) is too high", level)
		level = i.GetLevels()
	}

	return level
}
