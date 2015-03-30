package openjpeg

// #cgo LDFLAGS: -lopenjp2
// #include "handlers.h"
import "C"

import (
	"errors"
	"fmt"
)

func finalizer(i *JP2Image) {
	i.CleanupResources()
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
