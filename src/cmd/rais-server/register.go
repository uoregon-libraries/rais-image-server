package main

import (
	"net/url"
	"rais/src/img"
	"rais/src/openjpeg"
	"rais/src/plugins"
)

// decodeJP2 is the last decoder function we try, after any plugins have been
// tried, so we don't actually care about the URL - we just try it and see what
// happens.
func decodeJP2(s img.Streamer) (img.DecodeFunc, error) {
	return func() (img.Decoder, error) { return openjpeg.NewJP2Image(s) }, nil
}

// streamFiles is the last, and default, streamer for RAIS to try
func streamFiles(u *url.URL) (img.Streamer, error) {
	if u.Scheme != "file" {
		return nil, plugins.ErrSkipped
	}
	return img.NewFileStream(u.Path)
}
