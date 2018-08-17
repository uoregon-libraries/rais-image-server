package main

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path"
	"rais/src/iiif"
	"rais/src/transform"
	"strings"
)

// Custom errors an image read/transform operation could return
var (
	ErrImageDoesNotExist = errors.New("image file does not exist")
	ErrInvalidFiletype   = errors.New("invalid or unknown file type")
	ErrDecodeImage       = NewError("unable to decode image", 500)
	ErrBadImageFile      = errors.New("unable to read image")
)

// IIIFImageDecoder defines an interface for reading images in a generic way.  It's
// heavily biased toward the way we've had to do our JP2 images since they're
// the more unusual use-case.
type IIIFImageDecoder interface {
	DecodeImage() (image.Image, error)
	GetWidth() int
	GetHeight() int
	GetTileWidth() int
	GetTileHeight() int
	GetLevels() int
	SetCrop(image.Rectangle)
	SetResizeWH(int, int)
}

// ImageResource wraps a decoder, IIIF ID, and the path to the image
type ImageResource struct {
	Decoder  IIIFImageDecoder
	ID       iiif.ID
	FilePath string
}

// NewImageResource initializes and returns an ImageResource for the given id
// and path.  If the path doesn't resolve to a valid file, or resolves to a
// file type that isn't supported, an error is returned.  File type is
// determined by extension, so images will need standard extensions in order to
// work.
func NewImageResource(id iiif.ID, filepath string) (*ImageResource, error) {
	var err error

	// First, does the file exist?
	if _, err = os.Stat(filepath); err != nil {
		Logger.Infof("Image does not exist: %#v", filepath)
		return nil, ErrImageDoesNotExist
	}

	// File exists - is its extension registered?
	newDecoder, ok := ExtDecoders[strings.ToLower(path.Ext(filepath))]
	if !ok {
		Logger.Errorf("Image type unknown / invalid: %#v", filepath)
		return nil, ErrInvalidFiletype
	}

	// We have a decoder for the file type - attempt to instantiate it
	d, err := newDecoder(filepath)
	if err != nil {
		Logger.Errorf("Unable to read image %#v: %s", filepath, err)
		return nil, ErrBadImageFile
	}

	img := &ImageResource{ID: id, Decoder: d, FilePath: filepath}
	return img, nil
}

// Apply runs all image manipulation operations described by the IIIF URL, and
// returns an image.Image ready for encoding to the client
func (res *ImageResource) Apply(u *iiif.URL, max constraint) (image.Image, *HandlerError) {
	// Crop and resize have to be prepared before we can decode
	w, h := res.Decoder.GetWidth(), res.Decoder.GetHeight()
	crop := u.Region.GetCrop(w, h)
	scale := u.Size.GetResize(crop)

	// Determine the final image output dimensions to test size constraints
	sw, sh := scale.Dx(), scale.Dy()
	if u.Rotation.Degrees == 90 || u.Rotation.Degrees == 270 {
		sw, sh = sh, sw
	}
	if max.smallerThanAny(sw, sh) {
		return nil, NewError("requested image size exceeds server maximums", 501)
	}

	res.Decoder.SetCrop(crop)
	res.Decoder.SetResizeWH(scale.Dx(), scale.Dy())

	img, err := res.Decoder.DecodeImage()
	if err != nil {
		Logger.Errorf("Unable to decode image: %s", err)
		return nil, ErrDecodeImage
	}

	if u.Rotation.Mirror || u.Rotation.Degrees != 0 {
		img = rotate(img, u.Rotation)
	}

	// Unless I'm missing something, QColor doesn't actually change an image -
	// e.g., if it's already color, nothing happens.  If it's grayscale, there's
	// nothing to do (obviously we shouldn't report it, but oh well)
	switch u.Quality {
	case iiif.QGray:
		img = grayscale(img)
	case iiif.QBitonal:
		img = bitonal(img)
	}

	return img, nil
}

func rotate(img image.Image, rot iiif.Rotation) image.Image {
	var r transform.Rotator
	switch img0 := img.(type) {
	case *image.Gray:
		r = &transform.GrayRotator{Img: img0}
	case *image.RGBA:
		r = &transform.RGBARotator{Img: img0}
	}

	if rot.Mirror {
		r.Mirror()
	}

	switch rot.Degrees {
	case 90:
		r.Rotate90()
	case 180:
		r.Rotate180()
	case 270:
		r.Rotate270()
	}

	return r.Image()
}

func grayscale(img image.Image) image.Image {
	cm := img.ColorModel()
	if cm == color.GrayModel || cm == color.Gray16Model {
		return img
	}

	b := img.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, b, img, b.Min, draw.Src)
	return dst
}

func bitonal(img image.Image) image.Image {
	// First turn the image into 8-bit grayscale for easier manipulation
	imgGray := grayscale(img).(*image.Gray)
	b := imgGray.Bounds()
	imgBitonal := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	for i, pixel := range imgGray.Pix {
		if pixel > 190 {
			imgBitonal.Pix[i] = 255
		}
	}

	return imgBitonal
}
