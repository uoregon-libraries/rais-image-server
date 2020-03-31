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

// fileStreamReader is the last, and default, streamer for RAIS to try... it's
// also our last, best chance for peace.
func fileStreamReader(u *url.URL) (img.OpenStreamFunc, error) {
	if u.Scheme != "file" {
		return nil, plugins.ErrSkipped
	}

	return func() (img.Streamer, error) { return img.NewFileStream(u.Path) }, nil
}
