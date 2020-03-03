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

// DecodeFunc is a function which takes a file path and returns a Decoder and
// optionally an error.  If the error is ErrNotHandled, the decode function is
// stating that the filetype (or some other data inferred from the id) can't be
// handled by this decoder.
//
// TODO: this needs to be changed systemically to allow for a iiif ID rather
// than a path.  ID-to-stream lookups need to be implemented, not ID-to-path.
type DecodeFunc func(string) (Decoder, error)

// fns is our internal list of registered decoder functions
var decodeFuncs []DecodeFunc

// RegisterDecoder adds a decoder to the internal list of registered decoders.
// Images we want to decode will be run through each DecodeFn until one returns
// a Decoder and nil error.
func RegisterDecoder(fn DecodeFunc) {
	decodeFuncs = append(decodeFuncs, fn)
}
