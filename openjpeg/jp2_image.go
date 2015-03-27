package openjpeg

// #cgo LDFLAGS: -lopenjp2
// #include "handlers.h"
import "C"

import (
	"errors"
	"fmt"
	"image"
	"reflect"
	"runtime"
	"unsafe"
)

type JP2Image struct {
	filename                  string
	stream                    *C.opj_stream_t
	codec                     *C.opj_codec_t
	image                     *C.opj_image_t
	decodeWidth, decodeHeight int
	decodeArea                image.Rectangle
	crop, resize              bool
}

func finalizer(i *JP2Image) {
	i.CleanupResources()
}

func NewJP2Image(filename string) *JP2Image {
	i := &JP2Image{filename: filename}
	runtime.SetFinalizer(i, finalizer)
	return i
}

func (i *JP2Image) initializeStream() error {
	i.stream = C.opj_stream_create_default_file_stream(C.CString(i.filename), 1)
	if i.stream == nil {
		return errors.New(fmt.Sprintf("Failed to create stream in %#v", i.filename))
	}
	return nil
}

func (i *JP2Image) initializeCodec() error {
	i.codec = C.opj_create_decompress(C.OPJ_CODEC_JP2)

	var parameters C.opj_dparameters_t
	C.opj_set_default_decoder_parameters(&parameters)

	C.set_handlers(i.codec)
	if C.opj_setup_decoder(i.codec, &parameters) == C.OPJ_FALSE {
		return errors.New("failed to setup decoder")
	}
	return nil
}

func (i *JP2Image) SetResize(width, height int) {
	i.decodeWidth = width
	i.decodeHeight = height
	i.resize = true
}

func (i *JP2Image) SetCrop(r image.Rectangle) {
	i.decodeArea = r
	i.crop = true
}

func (i *JP2Image) RawImage() (*RawImage, error) {
	// If we want to resize, but not crop, we have to read the header (to get
	// dimensions), figure out progression level, and throw out all resources so
	// we can re-initialize with the right progression level
	if i.resize && !i.crop {
		i.ReadHeader()
		r, err := i.Dimensions()
		if err != nil {
			return nil, err
		}
		i.SetCrop(r)
		i.CleanupResources()
	}

	// Get progression level if we're resizing and cropping
	if i.resize && i.crop {
		if err := i.initializeCodec(); err != nil {
			return nil, err
		}

		level := desired_progression_level(i.decodeArea, i.decodeWidth, i.decodeHeight)
		goLog(6, fmt.Sprintf("desired level: %d", level))

		if C.opj_set_decoded_resolution_factor(i.codec, C.OPJ_UINT32(level)) == C.OPJ_FALSE {
			return nil, errors.New("failed to set decode resolution factor")
		}
	}

	if err := i.ReadHeader(); err != nil {
		return nil, err
	}

	goLog(6, fmt.Sprintf("num comps: %d", i.image.numcomps))
	goLog(6, fmt.Sprintf("x0: %d, x1: %d, y0: %d, y1: %d", i.image.x0, i.image.x1, i.image.y0, i.image.y1))

	// Setting decode area has to happen *after* reading the header / image data
	r := i.decodeArea
	if C.opj_set_decode_area(i.codec, i.image, C.OPJ_INT32(r.Min.X), C.OPJ_INT32(r.Min.Y), C.OPJ_INT32(r.Max.X), C.OPJ_INT32(r.Max.Y)) == C.OPJ_FALSE {
		return nil, errors.New("failed to set the decoded area")
	}

	// Decode the JP2 into the image stream
	if C.opj_decode(i.codec, i.stream, i.image) == C.OPJ_FALSE {
		return nil, errors.New("failed to decode image")
	}
	if C.opj_end_decompress(i.codec, i.stream) == C.OPJ_FALSE {
		return nil, errors.New("failed to decode image")
	}

	var comps []C.opj_image_comp_t
	compsSlice := (*reflect.SliceHeader)((unsafe.Pointer(&comps)))
	compsSlice.Cap = int(i.image.numcomps)
	compsSlice.Len = int(i.image.numcomps)
	compsSlice.Data = uintptr(unsafe.Pointer(i.image.comps))

	bounds := image.Rect(0, 0, int(comps[0].w), int(comps[0].h))

	var data []int32
	dataSlice := (*reflect.SliceHeader)((unsafe.Pointer(&data)))
	dataSlice.Cap = bounds.Dx() * bounds.Dy()
	dataSlice.Len = bounds.Dx() * bounds.Dy()
	dataSlice.Data = uintptr(unsafe.Pointer(comps[0].data))

	return &RawImage{data, bounds, bounds.Dx()}, nil
}

func (i *JP2Image) ReadHeader() error {
	if i.stream == nil {
		if err := i.initializeStream(); err != nil {
			return err
		}
	}

	if i.codec == nil {
		if err := i.initializeCodec(); err != nil {
			return err
		}
	}

	if C.opj_read_header(i.stream, i.codec, &i.image) == C.OPJ_FALSE {
		return errors.New("failed to read the header")
	}
	return nil
}

func (i *JP2Image) Dimensions() (r image.Rectangle, err error) {
	if i.image == nil {
		if err = i.ReadHeader(); err != nil {
			return
		}
	}
	r = image.Rect(int(i.image.x0), int(i.image.y0), int(i.image.x1), int(i.image.y1))
	return
}

func (i *JP2Image) CleanupResources() {
	if i.stream != nil {
		C.opj_stream_destroy(i.stream)
		i.stream = nil
	}

	if i.codec != nil {
		C.opj_destroy_codec(i.codec)
		i.codec = nil
	}

	if i.image != nil {
		C.opj_image_destroy(i.image)
		i.image = nil
	}
}
