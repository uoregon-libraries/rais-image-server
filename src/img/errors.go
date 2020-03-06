package img

// imgError is just a glorified string so we can have error constants
type imgError string

func (re imgError) Error() string {
	return string(re)
}

// Custom errors an image read/transform operation could return
const (
	ErrDoesNotExist           imgError = "image file does not exist"
	ErrInvalidFiletype        imgError = "invalid or unknown file type"
	ErrDimensionsExceedLimits imgError = "requested image size exceeds server maximums"
	ErrNotStreamable          imgError = "no registered streamers"
)
