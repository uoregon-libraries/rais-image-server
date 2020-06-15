package img

import (
	"image"
)

// Decoder defines an interface for reading images in a generic way.  It's
// heavily biased toward the way we've had to do our JP2 images since they're
// the more unusual use-case.
type Decoder interface {
	DecodeImage() (image.Image, error)
	GetWidth() int
	GetHeight() int
	GetTileWidth() int
	GetTileHeight() int
	GetLevels() int
	SetCrop(image.Rectangle)
	SetResizeWH(int, int)
}

// DecodeHandler is a function which takes a Streamer and returns a DecodeFunc and
// optionally an error.  If the error is ErrSkipped, the function is stating
// that it doesn't handle images the Streamer describes (typically just a brief
// check on the URL suffices, but a plugin could choose to read data from the
// streamer to get, e.g., a proper mime type).  A return with a nil error means
// the returned function should be used and searching is done.
type DecodeHandler func(Streamer) (DecodeFunc, error)

// DecodeFunc is the actual function which must be called for decoding its info
// / image data.  Since this will typically read from the stream immediately,
// it may return an error.  A handler is expected to hold onto its Streamer so
// its returned DecodeFunc doesn't have to take an unnecessary parameter.
type DecodeFunc func() (Decoder, error)

// decodeHandlers is our internal list of registered decoder functions
var decodeHandlers []DecodeHandler

// RegisterDecodeHandler adds a DecodeHandler to the internal list of
// registered handlers.  Images we want to decode will be run through each
// function until one returns a handler and nil error.
func RegisterDecodeHandler(fn DecodeHandler) {
	decodeHandlers = append(decodeHandlers, fn)
}
