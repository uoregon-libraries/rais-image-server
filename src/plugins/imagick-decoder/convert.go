package main

/*
#cgo pkg-config: MagickCore
#include <magick/MagickCore.h>
#include "magick.h"
*/
import "C"
import (
	"fmt"
	"image"
	"reflect"
	"unsafe"
)

// image returns a native Go image interface.  For now, this is always RGBA for
// simplicity, but it would be a good idea to use a gray image when it makes
// sense to improve performance and RAM usage.
func (i *Image) image(cimg *C.Image) (image.Image, error) {
	// Create and prep-for-freeing the exception
	exception := C.AcquireExceptionInfo()
	defer C.DestroyExceptionInfo(exception)

	img := image.NewRGBA(image.Rect(0, 0, i.decodeWidth, i.decodeHeight))

	area := i.decodeWidth * i.decodeHeight
	pixLen := area << 2
	pixels := make([]byte, pixLen)
	pi := reflect.ValueOf(pixels).Interface()
	ptr := unsafe.Pointer(&pixels[0])

	// Dimensions as C types
	w := C.size_t(i.decodeWidth)
	h := C.size_t(i.decodeHeight)

	var err = i.attemptExportRGBA(cimg, w, h, ptr, exception, 0)
	if err != nil {
		return nil, err
	}

	var ok bool
	img.Pix, ok = pi.([]uint8)
	if !ok {
		return nil, fmt.Errorf("unable to cast img.Pix to []uint8")
	}

	return img, nil
}

func (i *Image) attemptExportRGBA(cimg *C.Image, w, h C.size_t, ptr unsafe.Pointer, ex *C.ExceptionInfo, tries int) (err error) {
	defer func() {
		if x := recover(); x != nil {
			if tries < 3 {
				l.Warnf("Error trying to decode from ImageMagick (trying again): %s", x)
				_ = i.attemptExportRGBA(cimg, w, h, ptr, ex, tries+1)
			} else {
				l.Errorf("Error trying to decode from ImageMagick: %s", x)
				err = fmt.Errorf("imagemagick failure: %s", x)
			}
		}
	}()

	C.ExportRGBA(cimg, w, h, ptr, ex)
	return err
}
