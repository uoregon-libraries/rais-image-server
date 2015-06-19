package main

import "github.com/uoregon-libraries/rais-image-server/magick"

func init() {
	extList := []string{".tif", ".tiff", ".png", ".jpg", "jpeg", ".gif"}
	for _, ext := range extList {
		RegisterDecoder(ext, decodeCommonFile)
	}
}

func decodeCommonFile(path string) (IIIFImage, error) {
	return magick.NewImage(path)
}
