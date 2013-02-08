package main

// #cgo LDFLAGS: -lopenjp2
// #include <stdio.h>
// #include <openjpeg-2.0/openjpeg.h>
import "C"

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"reflect"
	"runtime"
	"unsafe"
)

const MAX_PROGRESSION_LEVEL = uint(6)

func scaled_dimension(progression_level uint, dimension int) int {
	scale_factor := uint(2) << (progression_level - uint(1))
	return int(float32(dimension) / float32(scale_factor))
}

func desired_progression_level(r image.Rectangle, width, height int) uint {
	level := MAX_PROGRESSION_LEVEL
	for ; level > 1 && width > scaled_dimension(level, r.Dx()) && height > scaled_dimension(level, r.Dy()); level-- {
	}
	return level
}

func NewImageTile(filename string, r image.Rectangle, width, height int) (err error, tile *ImageTile) {
	fsrc := C.fopen(C.CString(filename), C.CString("rb"))
	if fsrc == nil {
		return fmt.Errorf("failed to open '%s' for reading", filename), nil
	}
	defer func() {
		if fsrc != nil {
			C.fclose(fsrc)
		}
	}()
	l_stream := C.opj_stream_create_default_file_stream(fsrc, 1)
	if l_stream == nil {
		return errors.New("failed to create stream"), nil
	}

	l_codec := C.opj_create_decompress(C.OPJ_CODEC_JP2)

	var parameters C.opj_dparameters_t
	C.opj_set_default_decoder_parameters(&parameters)
	level := desired_progression_level(r, width, height)
	log.Println("desired level:", level)
	//(parameters).cp_reduce = C.OPJ_UINT32(level)

	if err == nil && C.opj_setup_decoder(l_codec, &parameters) == C.OPJ_FALSE {
		err = errors.New("failed to setup decoder")
	}

	if err == nil && C.opj_set_decoded_resolution_factor(l_codec, C.OPJ_UINT32(level)) == C.OPJ_FALSE {
		err = errors.New("failed to set decode resolution factor")
	}

	var img *C.opj_image_t
	if err == nil && C.opj_read_header(l_stream, l_codec, &img) == C.OPJ_FALSE {
		err = errors.New("failed to read the header")
	}

	log.Println("num comps:", img.numcomps)
	log.Println("x0:", img.x0, "x1:", img.x1, "y0:", img.y0, "y1:", img.y1)

	if err == nil && C.opj_set_decode_area(l_codec, img, C.OPJ_INT32(r.Min.X), C.OPJ_INT32(r.Min.Y), C.OPJ_INT32(r.Max.X), C.OPJ_INT32(r.Max.Y)) == C.OPJ_FALSE {
		err = errors.New("failed to set the decoded area")
	}

	if err == nil && C.opj_decode(l_codec, l_stream, img) == C.OPJ_FALSE {
		err = errors.New("failed to decode image")
	}
	if err == nil && C.opj_end_decompress(l_codec, l_stream) == C.OPJ_FALSE {
		err = errors.New("failed to decode image")
	}

	C.opj_stream_destroy(l_stream)
	if l_codec != nil {
		C.opj_destroy_codec(l_codec)
	}

	if err == nil {
		var comps []C.opj_image_comp_t
		compsSlice := (*reflect.SliceHeader)((unsafe.Pointer(&comps)))
		compsSlice.Cap = int(img.numcomps)
		compsSlice.Len = int(img.numcomps)
		compsSlice.Data = uintptr(unsafe.Pointer(img.comps))

		bounds := image.Rect(0, 0, int(comps[0].w), int(comps[0].h))

		var data []int32
		dataSlice := (*reflect.SliceHeader)((unsafe.Pointer(&data)))
		dataSlice.Cap = bounds.Dx() * bounds.Dy()
		dataSlice.Len = bounds.Dx() * bounds.Dy()
		dataSlice.Data = uintptr(unsafe.Pointer(comps[0].data))

		tile = &ImageTile{data, bounds, bounds.Dx(), img}
		runtime.SetFinalizer(tile, func(it *ImageTile) {
			C.opj_image_destroy(it.img)
		})
	} else {
		C.opj_image_destroy(img)
	}
	return
}

type ImageTile struct {
	data   []int32
	bounds image.Rectangle
	stride int
	img    *C.opj_image_t
}

func (p *ImageTile) ColorModel() color.Model {
	return color.GrayModel
}

func (p *ImageTile) Bounds() image.Rectangle {
	return p.bounds
}

func (p *ImageTile) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.bounds)) {
		return color.Gray{}
	}
	index := p.PixOffset(x, y)
	return color.Gray{uint8(p.data[index])}
}

func (p *ImageTile) PixOffset(x, y int) int {
	return (y-p.bounds.Min.Y)*p.stride + (x-p.bounds.Min.X)*1
}
