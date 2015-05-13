package main

import (
	"errors"
	"fmt"
	"github.com/uoregon-libraries/rais-image-server/iiif"
	"github.com/uoregon-libraries/rais-image-server/openjpeg"
	"github.com/uoregon-libraries/rais-image-server/transform"
	"image"
	"log"
	"os"
	"path"
	"strings"
)

var (
	ErrImageDoesNotExist = errors.New("Image file does not exist")
	ErrInvalidFiletype   = errors.New("Invalid or unknown file type")
	ErrDecodeImage       = errors.New("Unable to decode image")
)

// IIIFImage defines an interface for reading images in a generic way.  It's
// heavily biased toward the way we've had to do our JP2 images since they're
// the more unusual use-case.
type IIIFImage interface {
	CleanupResources()
	DecodeImage() (image.Image, error)
	GetDimensions() (image.Rectangle, error)
	SetCrop(image.Rectangle)
	SetResizeWH(int, int)
	SetScale(float64)
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
		return nil, errors.New(fmt.Sprintf("Unable to read image %#v: %s", id, err))
	}

	img := &ImageResource{ID: id, Image: i, FilePath: filepath}
	return img, nil
}

// Apply runs all image manipulation operations described by the IIIF URL, and
// returns an image.Image ready for encoding to the client
func (res *ImageResource) Apply(u *iiif.URL) (image.Image, error) {
	// Crop and resize have to be prepared before we can decode
	if err := res.prepCrop(u.Region); err != nil {
		return nil, err
	}

	res.prepResize(u.Size)

	img, err := res.Image.DecodeImage()
	if err != nil {
		log.Println("Unable to decode image: ", err)
		return nil, ErrDecodeImage
	}

	if u.Rotation.Degrees != 0 {
		img = rotate(img, u.Rotation)
	}

	return img, nil
}

func (res *ImageResource) prepCrop(r iiif.Region) error {
	switch r.Type {
	case iiif.RTPixel:
		rect := image.Rect(int(r.X), int(r.Y), int(r.X+r.W), int(r.Y+r.H))
		res.Image.SetCrop(rect)
	case iiif.RTPercent:
		dim, err := res.Image.GetDimensions()
		if err != nil {
			return err
		}
		rect := image.Rect(
			int(r.X * float64(dim.Dx()) / 100.0),
			int(r.Y * float64(dim.Dy()) / 100.0),
			int((r.X+r.W) * float64(dim.Dx()) / 100.0),
			int((r.Y+r.H) * float64(dim.Dy()) / 100.0),
		)
		res.Image.SetCrop(rect)
	}

	return nil
}

func (res *ImageResource) prepResize(s iiif.Size) {
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
