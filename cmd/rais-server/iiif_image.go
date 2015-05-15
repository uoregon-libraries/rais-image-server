package main

import (
	"errors"
	"github.com/uoregon-libraries/rais-image-server/iiif"
	"github.com/uoregon-libraries/rais-image-server/openjpeg"
	"github.com/uoregon-libraries/rais-image-server/transform"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"path"
	"strings"
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
	SetScale(float64)
}

type ImageResource struct {
	Image      IIIFImage
	ID         iiif.ID
	FilePath   string
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

	// File exists - is it a valid filetype?
	var i IIIFImage
	fileExt := strings.ToLower(path.Ext(filepath))
	switch fileExt {
	case ".jp2":
		i, err = openjpeg.NewJP2Image(filepath)
	case ".tif", ".tiff", ".png", ".jpg", "jpeg", ".gif":
		i, err = NewSimpleImage(filepath)
	default:
		log.Printf("Image type unknown / invalid: %#v", filepath)
		return nil, ErrInvalidFiletype
	}

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

	if u.Rotation.Degrees != 0 {
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
	var crop image.Rectangle

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

	switch s.Type {
	case iiif.STScaleToWidth:
		res.Image.SetResizeWH(s.W, 0)
	case iiif.STScaleToHeight:
		res.Image.SetResizeWH(0, s.H)
	case iiif.STExact:
		res.Image.SetResizeWH(s.W, s.H)
	case iiif.STScalePercent:
		res.Image.SetScale(s.Percent / 100.0)
	}
}

func rotate(img image.Image, rot iiif.Rotation) image.Image {
	var r transform.Rotator
	switch img0 := img.(type) {
	case *image.Gray:
		r = transform.GrayRotator{img0}
	case *image.RGBA:
		r = transform.RGBARotator{img0}
	}

	switch rot.Degrees {
	case 90:
		img = r.Rotate90()
	case 180:
		img = r.Rotate180()
	case 270:
		img = r.Rotate270()
	}

	return img
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
