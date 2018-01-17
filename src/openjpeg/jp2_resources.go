package openjpeg

// #cgo pkg-config: libopenjp2
// #include <openjpeg.h>
// #include <stdlib.h>
// #include "handlers.h"
import "C"

import (
	"fmt"
	"unsafe"
)

// rawDecode runs the low-level operations necessary to actually get the
// desired tile/resized image
func (i *JP2Image) rawDecode() (jp2 *C.opj_image_t, err error) {
	// Setup the parameters for decode
	var parameters C.opj_dparameters_t
	C.opj_set_default_decoder_parameters(&parameters)

	// Calculate cp_reduce - this seems smarter to put in a parameter than to call an extra function
	parameters.cp_reduce = C.OPJ_UINT32(i.computeProgressionLevel())

	// Setup file stream
	stream, err := initializeStream(i.filename)
	if err != nil {
		return jp2, err
	}
	defer C.opj_stream_destroy(stream)

	// Create codec
	codec := C.opj_create_decompress(C.OPJ_CODEC_JP2)
	defer C.opj_destroy_codec(codec)

	// Connect our info/warning/error handlers
	C.set_handlers(codec)

	// Fill in codec configuration from parameters
	if C.opj_setup_decoder(codec, &parameters) == C.OPJ_FALSE {
		return jp2, fmt.Errorf("unable to setup decoder")
	}

	// Read the header to set up the image data
	if C.opj_read_header(stream, codec, &jp2) == C.OPJ_FALSE {
		return jp2, fmt.Errorf("failed to read the header")
	}

	Logger.Debugf("num comps: %d", jp2.numcomps)
	Logger.Debugf("x0: %d, x1: %d, y0: %d, y1: %d", jp2.x0, jp2.x1, jp2.y0, jp2.y1)

	// Set the decode area if it isn't the full image
	if i.decodeArea != i.srcRect {
		r := i.decodeArea
		if C.opj_set_decode_area(codec, jp2, C.OPJ_INT32(r.Min.X), C.OPJ_INT32(r.Min.Y), C.OPJ_INT32(r.Max.X), C.OPJ_INT32(r.Max.Y)) == C.OPJ_FALSE {
			return jp2, fmt.Errorf("failed to set the decoded area")
		}
	}

	// Decode the JP2 into the image stream
	if C.opj_decode(codec, stream, jp2) == C.OPJ_FALSE || C.opj_end_decompress(codec, stream) == C.OPJ_FALSE {
		return jp2, fmt.Errorf("failed to decode image")
	}

	return jp2, nil
}

func initializeStream(filename string) (*C.opj_stream_t, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	stream := C.opj_stream_create_default_file_stream(cFilename, 1)
	if stream == nil {
		return nil, fmt.Errorf("failed to create stream in %#v", filename)
	}
	return stream, nil
}
