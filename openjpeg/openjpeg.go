package openjpeg

// #cgo LDFLAGS: -lopenjp2
// #include "handlers.h"
import "C"

import (
	"errors"
	"image"
	"image/color"
	"reflect"
	"unsafe"
	"fmt"
)

func NewImageTile(jp2 *JP2Image, r image.Rectangle, width, height int) (tile *ImageTile, err error) {
	level := desired_progression_level(r, width, height)
	goLog(6, fmt.Sprintf("desired level: %d", level))
	//(parameters).cp_reduce = C.OPJ_UINT32(level)

	if err == nil && C.opj_set_decoded_resolution_factor(jp2.codec, C.OPJ_UINT32(level)) == C.OPJ_FALSE {
		err = errors.New("failed to set decode resolution factor")
	}

	// Header *must* be read after we determine the decode resolution factor
	err = jp2.ReadHeader()

	if err == nil {
		goLog(6, fmt.Sprintf("num comps: %d", jp2.image.numcomps))
		goLog(6, fmt.Sprintf("x0: %d, x1: %d, y0: %d, y1: %d", jp2.image.x0, jp2.image.x1, jp2.image.y0, jp2.image.y1))
	}

	if err == nil && C.opj_set_decode_area(jp2.codec, jp2.image, C.OPJ_INT32(r.Min.X), C.OPJ_INT32(r.Min.Y), C.OPJ_INT32(r.Max.X), C.OPJ_INT32(r.Max.Y)) == C.OPJ_FALSE {
		err = errors.New("failed to set the decoded area")
	}

	if err == nil && C.opj_decode(jp2.codec, jp2.stream, jp2.image) == C.OPJ_FALSE {
		err = errors.New("failed to decode image")
	}
	if err == nil && C.opj_end_decompress(jp2.codec, jp2.stream) == C.OPJ_FALSE {
		err = errors.New("failed to decode image")
	}

	if err == nil {
		var comps []C.opj_image_comp_t
		compsSlice := (*reflect.SliceHeader)((unsafe.Pointer(&comps)))
		compsSlice.Cap = int(jp2.image.numcomps)
		compsSlice.Len = int(jp2.image.numcomps)
		compsSlice.Data = uintptr(unsafe.Pointer(jp2.image.comps))

		bounds := image.Rect(0, 0, int(comps[0].w), int(comps[0].h))

		var data []int32
		dataSlice := (*reflect.SliceHeader)((unsafe.Pointer(&data)))
		dataSlice.Cap = bounds.Dx() * bounds.Dy()
		dataSlice.Len = bounds.Dx() * bounds.Dy()
		dataSlice.Data = uintptr(unsafe.Pointer(comps[0].data))

		tile = &ImageTile{data, bounds, bounds.Dx(), jp2.image}
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
