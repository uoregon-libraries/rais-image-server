package main

import (
	"rais/src/openjpeg"
)

func init() {
	RegisterDecoder(".jp2", decodeJP2)
}

func decodeJP2(path string) (IIIFImageDecoder, error) {
	return openjpeg.NewJP2Image(path)
}
