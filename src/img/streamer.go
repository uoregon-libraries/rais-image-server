package img

import (
	"io"
	"net/url"
)

// Streamer is an encapsulation of existence checking, reading, seeking, and
// closing so that we can implement image and info.json streaming from memory,
// a file, S3, etc.
type Streamer interface {
	Location() *url.URL // Location returns the URL to the object being streamed
	Exist() bool        // Exist checks if the object exists in the location defined by its URL
	io.ReadSeeker
	io.Closer
}

// StreamFunc is a function which takes an image URL and returns a Streamer and
// optionally an error.  If the error is ErrSkipped, the streamer function is
// stating that it doesn't handle images from the given URL (generally based on
// the URL scheme).  A return with a nil error means this streamer should be
// used and no other streamers should be searched.
type StreamFunc func(u *url.URL) (Streamer, error)

// streamFuncs is our internal list of registered streamer functions
var streamFuncs []StreamFunc

// RegisterStreamer adds a streamer to the internal list of registered
// streamers.  Image URLs will be run through each StreamFunc until one returns
// a Streamer and nil error.
func RegisterStreamer(fn StreamFunc) {
	streamFuncs = append(streamFuncs, fn)
}
