package magick

/*
#cgo pkg-config: MagickCore
#include <magick/MagickCore.h>
#include "magick.h"
*/
import "C"
import (
	"image"
	"runtime"
	"unsafe"
)

// Image implements IIIFImage for reading non-JP2 image types via image magick
// bindings.  Requires ImageMagick dev files to be installed.
//
// NOTE: To keep with the IIIFImage interface, we're not really using
// ImageMagick efficiently.  We don't let it rotate, change color depth,
// encode, etc.  We instead convert to a Go image, which is itself probably
// slow, and then let even less efficient code take over for those operations.
type Image struct {
	image        (*C.Image)
	imageInfo    (*C.ImageInfo)
	decodeWidth  int
	decodeHeight int
	decodeArea   image.Rectangle
}

// SetResizeWH sets the image to scale to the given width and height.  If one
// dimension is 0, the decoded image will preserve the aspect ratio while
// scaling to the non-zero dimension.
func (i *Image) SetResizeWH(width, height int) {
	i.decodeWidth = width
	i.decodeHeight = height
}

func (i *Image) SetCrop(r image.Rectangle) {
	i.decodeArea = r
}

func NewImage(filename string) (*Image, error) {
	exception := C.AcquireExceptionInfo()
	defer C.DestroyExceptionInfo(exception)

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	info := C.AcquireImageInfo()
	C.SetImageInfoFilename(info, cFilename)
	image := C.ReadImage(info, exception)
	if C.HasError(exception) == 1 {
		C.DestroyImageInfo(info)
		return nil, makeError(exception)
	}

	i := &Image{image: image, imageInfo: info}
	runtime.SetFinalizer(i, finalizer)
	return i, nil
}

func (i *Image) replace(newImg *C.Image) {
	i.CleanupImage()
	i.image = newImg
}

// Width returns the Width of the loaded image in pixels as an int
func (i *Image) GetWidth() int {
	return (int)(i.image.columns)
}

// Height returns the Height of the loaded image in pixels as an int
func (i *Image) GetHeight() int {
	return (int)(i.image.rows)
}

func (i *Image) doResize(w, h int) error {
	exception := C.AcquireExceptionInfo()
	defer C.DestroyExceptionInfo(exception)

	newImg := C.Resize(i.image, C.size_t(w), C.size_t(h), exception)
	if C.HasError(exception) == 1 {
		return makeError(exception)
	}

	i.replace(newImg)
	return nil
}

func (i *Image) doCrop(r image.Rectangle) error {
	exception := C.AcquireExceptionInfo()
	defer C.DestroyExceptionInfo(exception)

	var ri = C.MakeRectangle(C.int(r.Min.X), C.int(r.Min.Y), C.int(r.Dx()), C.int(r.Dy()))
	newImg := C.CropImage(i.image, &ri, exception)
	if C.HasError(exception) == 1 {
		return makeError(exception)
	}

	i.replace(newImg)
	return nil
}

// DecodeImage returns an image.Image that holds the decoded image data,
// resized and cropped if resizing or cropping was requested.  Both cropping
// and resizing happen here due to the nature of openjpeg and our desire to
// keep this API consistent with the jp2 api.
func (i *Image) DecodeImage() (image.Image, error) {
	w, h := i.GetWidth(), i.GetHeight()
	if i.decodeArea == image.ZR {
		i.decodeArea = image.Rect(0, 0, w, h)
	}

	if i.decodeWidth == 0 && i.decodeHeight == 0 {
		i.decodeWidth = w
		i.decodeHeight = h
	}

	if i.decodeWidth == 0 || i.decodeHeight == 0 {
		srcW64 := float64(i.GetWidth())
		srcH64 := float64(i.GetHeight())
		h64 := float64(i.decodeHeight)
		w64 := float64(i.decodeWidth)

		if w64 == 0 {
			scale := h64 / srcH64
			i.decodeWidth = int(scale * srcW64)
		}
		if h64 == 0 {
			scale := w64 / srcW64
			i.decodeHeight = int(scale * srcH64)
		}
	}

	// Crop if decode area isn't the same as the full image
	if i.decodeArea != image.Rect(0, 0, w, h) {
		err := i.doCrop(i.decodeArea)
		if err != nil {
			return nil, err
		}
	}

	if i.decodeWidth != i.decodeArea.Dx() || i.decodeHeight != i.decodeArea.Dy() {
		err := i.doResize(i.decodeWidth, i.decodeHeight)
		if err != nil {
			return nil, err
		}
	}

	return i.Image()
}
