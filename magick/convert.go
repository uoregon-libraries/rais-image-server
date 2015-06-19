package magick

/*
#cgo pkg-config: MagickCore
#include <magick/MagickCore.h>
#include "magick.h"
*/
import "C"
import (
	"image"
)

// Image returns a native Go image interface
func (i *Image) Image() (image.Image, error) {
	return image.NewRGBA(image.Rect(0, 0, i.decodeWidth, i.decodeHeight)), nil
}
