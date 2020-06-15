package img

import (
	"io"
	"net/url"
	"time"
)

// Streamer is an encapsulation of basic metadata checking, reading, seeking,
// and closing so that we can implement image and info.json streaming from
// memory, a file, S3, etc.
type Streamer interface {
	Location() *url.URL // Location returns the URL to the object being streamed
	Size() int64        // Size in bytes of the stream data
	ModTime() time.Time // When the data was last modified
	io.ReadSeeker
	io.Closer
}

// StreamReader is a function which takes a URL and returns an OpenStreamFunc
// and optionally an error.  The error should generally be nil (success) or
// ErrSkipped (the reader doesn't handle the given URL).  The returned function
// must be bound to the URL to avoid passing the URL around extra times, or
// worse, passing the wrong URL into an OpenStreamFunc that won't be able to
// handle it.
type StreamReader func(*url.URL) (OpenStreamFunc, error)

// OpenStreamFunc is the function which actually returns a Streamer (ready for
// use) or else an error.
type OpenStreamFunc func() (Streamer, error)

// streamFuncs is our internal list of registered streamer functions
var streamReaders []StreamReader

// RegisterStreamReader adds a reader to the internal list of registered
// readers.  Image URLs will be run through each reader until one returns an
// OpenStreamFunc and nil error.
func RegisterStreamReader(fn StreamReader) {
	streamReaders = append(streamReaders, fn)
}
