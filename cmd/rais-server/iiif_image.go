package main

import (
	"errors"
	"fmt"
	"github.com/uoregon-libraries/rais-image-server/iiif"
	"github.com/uoregon-libraries/rais-image-server/openjpeg"
	"image"
	"log"
	"os"
	"path"
)

var (
	ErrImageDoesNotExist = errors.New("Image file does not exist")
	ErrInvalidFiletype   = errors.New("Invalid or unknown file type")
)

// IIIFImage defines an interface for reading images in a generic way.  It's
// heavily biased toward the way we've had to do our JP2 images since they're
// the more unusual use-case.
type IIIFImage interface {
	CleanupResources()
	DecodeImage()            (image.Image, error)
	GetDimensions()          (image.Rectangle, error)
	SetCrop(image.Rectangle)
	SetResizeWH(int, int)
	SetScale(float64)
}

type ImageResource struct {
	Image    IIIFImage
	ID       iiif.ID
	FilePath string
}

// Initializes and returns an IIIFImage for the given id and path.  If the path
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
	fileExt := path.Ext(filepath)
	switch fileExt {
	case ".jp2":
		i, err = openjpeg.NewJP2Image(filepath)
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
