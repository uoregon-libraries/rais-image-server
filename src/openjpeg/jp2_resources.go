package openjpeg

// #cgo pkg-config: libopenjp2
// #include <openjpeg.h>
// #include <stdlib.h>
// #include "handlers.h"
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

func finalizer(i *JP2Image) {
	i.CleanupResources()
}

func (i *JP2Image) initializeStream() error {
	if i.stream != nil {
		return nil
	}

	cFilename := C.CString(i.filename)
	defer C.free(unsafe.Pointer(cFilename))

	i.stream = C.opj_stream_create_default_file_stream(cFilename, 1)
	if i.stream == nil {
		return errors.New(fmt.Sprintf("Failed to create stream in %#v", i.filename))
	}
	return nil
}

func (i *JP2Image) initializeCodec() error {
	if i.codec != nil {
		return nil
	}

	i.codec = C.opj_create_decompress(C.OPJ_CODEC_JP2)

	var parameters C.opj_dparameters_t
	C.opj_set_default_decoder_parameters(&parameters)

	C.set_handlers(i.codec)
	if C.opj_setup_decoder(i.codec, &parameters) == C.OPJ_FALSE {
		return errors.New("failed to setup decoder")
	}
	return nil
}

func (i *JP2Image) cleanupStream() {
	if i.stream != nil {
		C.opj_stream_destroy(i.stream)
		i.stream = nil
	}
}

func (i *JP2Image) cleanupCodec() {
	if i.codec != nil {
		C.opj_destroy_codec(i.codec)
		i.codec = nil
	}
}

func (i *JP2Image) cleanupImage() {
	if i.image != nil {
		C.opj_image_destroy(i.image)
		i.image = nil
	}
}

func (i *JP2Image) CleanupResources() {
	i.cleanupStream()
	i.cleanupCodec()
	i.cleanupImage()
}
