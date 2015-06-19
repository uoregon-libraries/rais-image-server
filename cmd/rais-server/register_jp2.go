//+build jp2

package main

import (
	"github.com/uoregon-libraries/rais-image-server/openjpeg"
)

func init() {
	RegisterDecoder(".jp2", decodeJP2)
}

func decodeJP2(path string) (IIIFImage, error) {
	return openjpeg.NewJP2Image(path)
}
