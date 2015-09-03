package main

import (
	"errors"
	"iiif"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"path"
	"strings"
	"transform"
)

var (
	ErrImageDoesNotExist = errors.New("Image file does not exist")
	ErrInvalidFiletype   = errors.New("Invalid or unknown file type")
	ErrDecodeImage       = errors.New("Unable to decode image")
	ErrBadImageFile      = errors.New("Unable to read image")
)

// IIIFImage defines an interface for reading images in a generic way.  It's
// heavily biased toward the way we've had to do our JP2 images since they're
// the more unusual use-case.
type IIIFImage interface {
	DecodeImage() (image.Image, error)
	GetWidth() int
	GetHeight() int
	SetCrop(image.Rectangle)
	SetResizeWH(int, int)
}

type ImageResource struct {
	Image    IIIFImage
	ID       iiif.ID
	FilePath string
}

// Initializes and returns an ImageResource for the given id and path.  If the path
// doesn't resolve to a valid file, or resolves to a file type that isn't
// supported, an error is returned.  File type is determined by extension, so
// images will need standard extensions in order to work.
func NewImageResource(id iiif.ID, filepath string) (*ImageResource, error) {
	var err error

	// First, does the file exist?
	if _, err = os.Stat(filepath); err != nil {
		log.Printf("Image does not exist: %#v", filepath)
		return nil, ErrImageDoesNotExist
	}

	// File exists - is its extension registered?
	decoder, ok := ExtDecoders[strings.ToLower(path.Ext(filepath))]
	if !ok {
		log.Printf("Image type unknown / invalid: %#v", filepath)
		return nil, ErrInvalidFiletype
	}

	// We have a decoder for the file type - attempt to decode
	i, err := decoder(filepath)
	if err != nil {
		log.Printf("Unable to read image %#v: %s", filepath)
		return nil, ErrBadImageFile
	}

	img := &ImageResource{ID: id, Image: i, FilePath: filepath}
	return img, nil
}

// Apply runs all image manipulation operations described by the IIIF URL, and
// returns an image.Image ready for encoding to the client
func (res *ImageResource) Apply(u *iiif.URL) (image.Image, error) {
	// Crop and resize have to be prepared before we can decode
	res.prep(u.Region, u.Size)

	img, err := res.Image.DecodeImage()
	if err != nil {
		log.Println("Unable to decode image: ", err)
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

func (res *ImageResource) prep(r iiif.Region, s iiif.Size) {
	w, h := res.Image.GetWidth(), res.Image.GetHeight()
	crop := image.Rect(0, 0, w, h)

	switch r.Type {
	case iiif.RTPixel:
		crop = image.Rect(int(r.X), int(r.Y), int(r.X+r.W), int(r.Y+r.H))
	case iiif.RTPercent:
		crop = image.Rect(
			int(r.X*float64(w)/100.0),
			int(r.Y*float64(h)/100.0),
			int((r.X+r.W)*float64(w)/100.0),
			int((r.Y+r.H)*float64(h)/100.0),
		)
	}
	res.Image.SetCrop(crop)

	w, h = crop.Dx(), crop.Dy()
	switch s.Type {
	case iiif.STScaleToWidth:
		w, h = s.W, 0
	case iiif.STScaleToHeight:
		w, h = 0, s.H
	case iiif.STExact:
		w, h = s.W, s.H
	case iiif.STBestFit:
		w, h = res.getBestFit(w, h, s)
	case iiif.STScalePercent:
		w = int(float64(crop.Dx()) * s.Percent / 100.0)
		h = int(float64(crop.Dy()) * s.Percent / 100.0)
	}
	res.Image.SetResizeWH(w, h)
}

// Preserving the aspect ratio, determines the proper scaling factor to get
// width and height adjusted to fit within the width and height of the desired
// size operation
func (res *ImageResource) getBestFit(width, height int, s iiif.Size) (int, int) {
	fW, fH, fsW, fsH := float64(width), float64(height), float64(s.W), float64(s.H)
	sf := fsW / fW
	if sf*fH > fsH {
		sf = fsH / fH
	}
	return int(sf * fW), int(sf * fH)
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
