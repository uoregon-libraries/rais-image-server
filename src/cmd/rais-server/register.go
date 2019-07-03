package main

import (
	"path/filepath"
	"rais/src/img"
	"rais/src/magick"
	"rais/src/openjpeg"
)

func decodeJP2(path string) (img.Decoder, error) {
	if filepath.Ext(path) == ".jp2" {
		return openjpeg.NewJP2Image(path)
	}
	return nil, img.ErrNotHandled
}

func decodeCommonFile(path string) (img.Decoder, error) {
	switch filepath.Ext(path) {
	case ".tif", ".tiff", ".png", ".jpg", "jpeg", ".gif":
		return magick.NewImage(path)
	default:
		return nil, img.ErrNotHandled
	}
}
