package main

import (
	"errors"
	"github.com/uoregon-libraries/rais-image-server/iiif"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

var ErrInvalidEncodeFormat = errors.New("Unable to encode: unsupported format")

func EncodeImage(w io.Writer, img image.Image, format iiif.Format) error {
	switch format {
	case iiif.FmtJPG:
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 80})
	case iiif.FmtPNG:
		return png.Encode(w, img)
	case iiif.FmtGIF:
		return gif.Encode(w, img, &gif.Options{NumColors: 256})
	}

	return ErrInvalidEncodeFormat
}
