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
	file         *os.File
	conf         image.Config
	decodeWidth  int
	decodeHeight int
	decodeArea   image.Rectangle
	resize       bool
	crop         bool
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
	i.resize = false
	i.crop = false

	return i, nil
}

// SetScale sets the image to scale by the given multiplier, typically a
// percentage from 0 to 1.  This is mutually exclusive with resizing by a set
// width/height value.
func (i *SimpleImage) SetScale(m float64) {
	i.resize = true
	i.decodeWidth = int(float64(i.conf.Width) * m)
	i.decodeHeight = int(float64(i.conf.Height) * m)
}

// SetResizeWH sets the image to scale to the given width and height.  If one
// dimension is 0, the decoded image will preserve the aspect ratio while
// scaling to the non-zero dimension.
func (i *SimpleImage) SetResizeWH(width, height int) {
	i.resize = true
	i.decodeWidth = width
	i.decodeHeight = height
}

func (i *SimpleImage) SetCrop(r image.Rectangle) {
	i.crop = true
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

	if i.crop {
		srcB := img.Bounds()
		dstB := i.decodeArea
		dst := image.NewRGBA(image.Rect(0, 0, dstB.Dx(), dstB.Dy()))
		draw.Draw(dst, srcB, img, dstB.Min, draw.Src)
		img = dst
	}

	if i.resize {
		img = resize.Resize(uint(i.decodeWidth), uint(i.decodeHeight), img, resize.Bilinear)
	}

	return img, nil
}

// GetDimensions returns the config data as a rectangle
func (i *SimpleImage) GetDimensions() (image.Rectangle, error) {
	return image.Rect(0, 0, i.conf.Width, i.conf.Height), nil
}
