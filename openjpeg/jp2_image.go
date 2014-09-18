package openjpeg

// #cgo LDFLAGS: -lopenjp2
// #include "handlers.h"
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"image"
)

type JP2Image struct {
	filename string
	stream *C.opj_stream_t
	codec *C.opj_codec_t
	image *C.opj_image_t
}

func finalizer(i *JP2Image) {
	if i.stream != nil {
		C.opj_stream_destroy_v3(i.stream)
	}

	if i.codec != nil {
		C.opj_destroy_codec(i.codec)
	}

	if i.image != nil {
		C.opj_image_destroy(i.image)
	}
}

func NewJP2Image(filename string) (*JP2Image, error) {
	i := &JP2Image{filename: filename}
	runtime.SetFinalizer(i, finalizer)

	if err := i.initializeStream(); err != nil {
		return nil, err
	}
	if err := i.initializeCodec(); err != nil {
		return nil, err
	}

	return i, nil
}

func (i *JP2Image) initializeStream() error {
	i.stream = C.opj_stream_create_default_file_stream_v3(C.CString(i.filename), 1)
	if (i.stream == nil) {
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

func (i *JP2Image) ReadHeader() error {
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
