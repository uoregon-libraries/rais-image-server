package main

import (
	"github.com/nfnt/resize"
	_ "golang.org/x/image/tiff"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// SimpleImage implements IIIFImage for reading non-JP2 image types.  These can
// only handle a basic "read it all, then crop/resize" flow, and thus should
// have very careful load testing.
type SimpleImage struct {
	file            *os.File
	conf            image.Config
	decodeWidth     int
	decodeHeight    int
	scaleFactor     float64
	decodeArea      image.Rectangle
}

func NewSimpleImage(filename string) (*SimpleImage, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	i := &SimpleImage{file: file}
	i.conf, _, err = image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}
	i.file.Seek(0, 0)

	// Default to full size, no crop
	i.decodeWidth = i.conf.Width
	i.decodeHeight = i.conf.Height
	i.decodeArea = image.Rect(0, 0, i.decodeWidth, i.decodeHeight)

	return i, nil
}

// SetResizeWH sets the image to scale to the given width and height.  If one
// dimension is 0, the decoded image will preserve the aspect ratio while
// scaling to the non-zero dimension.
func (i *SimpleImage) SetResizeWH(width, height int) {
	i.decodeWidth = width
	i.decodeHeight = height
}

func (i *SimpleImage) SetCrop(r image.Rectangle) {
	i.decodeArea = r
}

// DecodeImage returns an image.Image that holds the decoded image data,
// resized and cropped if resizing or cropping was requested.  Both cropping
// and resizing happen here due to the nature of openjpeg, so SetScale,
// SetResizeWH, and SetCrop must be called before this function.
func (i *SimpleImage) DecodeImage() (image.Image, error) {
	img, _, err := image.Decode(i.file)
	if err != nil {
		return nil, err
	}

	// Draw a new image of the requested size if the decode area isn't the same
	// rectangle as the source image
	if i.decodeArea != img.Bounds() {
		srcB := img.Bounds()
		dstB := i.decodeArea
		dst := image.NewRGBA(image.Rect(0, 0, dstB.Dx(), dstB.Dy()))
		draw.Draw(dst, srcB, img, dstB.Min, draw.Src)
		img = dst
	}

	if i.decodeWidth != i.decodeArea.Dx() || i.decodeHeight != i.decodeArea.Dy() {
		img = resize.Resize(uint(i.decodeWidth), uint(i.decodeHeight), img, resize.Bilinear)
	}

	return img, nil
}

// GetWidth returns the image width
func (i *SimpleImage) GetWidth() int {
	return i.conf.Width
}

// GetHeight returns the image height
func (i *SimpleImage) GetHeight() int {
	return i.conf.Height
}
