package main

import (
	"rais/src/img"
	"rais/src/openjpeg"
)

// decodeJP2 is the last decoder function we try, after any plugins have been
// tried, so we don't actually care about the URL - we just try it and see what
// happens.
func decodeJP2(path string) (img.DecodeFunc, error) {
	return func() (img.Decoder, error) { return openjpeg.NewJP2Image(path) }, nil
}
