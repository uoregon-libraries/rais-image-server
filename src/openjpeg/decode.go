package openjpeg

// #cgo pkg-config: libopenjp2
// #include <openjpeg.h>
import "C"
import (
	"errors"
	"image"
	"reflect"
	"unsafe"
)

type opjp2 struct {
	comps  []C.opj_image_comp_t
	width  int
	height int
	bounds image.Rectangle
	bpc    uint8
}

func newOpjp2(comps []C.opj_image_comp_t, bpc uint8) (*opjp2, error) {
	if bpc != 8 && bpc != 16 {
		return nil, errors.New("bit depth must be 8 or 16")
	}

	var j = &opjp2{comps: comps}
	j.width = int(comps[0].w)
	j.height = int(comps[0].h)
	j.bounds = image.Rect(0, 0, j.width, j.height)
	j.bpc = bpc

	return j, nil
}

func (j *opjp2) decode() (image.Image, error) {
	var gray = len(j.comps) < 3
	var i image.Image
	switch {
	case j.bpc == 8 && gray:
		i = j.decodeGray8()
	case j.bpc == 8 && !gray:
		i = j.decodeRGB8()
	case j.bpc == 16 && gray:
		i = j.decodeGray16()
	case j.bpc == 16 && !gray:
		i = j.decodeRGB16()
	}

	return i, nil
}

func (j *opjp2) decodeGray8() image.Image {
	return &image.Gray{Pix: jp2ComponentData8(j.comps[0]), Stride: j.width, Rect: j.bounds}
}

func (j *opjp2) decodeGray16() image.Image {
	return &image.Gray16{Pix: jp2ComponentData16(j.comps[0]), Stride: j.width << 1, Rect: j.bounds}
}

func (j *opjp2) decodeRGB8() image.Image {
	var area = j.width * j.height
	var bytes = area << 2
	var realData = make([]uint8, bytes)

	var red = jp2ComponentData8(j.comps[0])
	var green = jp2ComponentData8(j.comps[1])
	var blue = jp2ComponentData8(j.comps[2])

	var offset = 0
	for i := 0; i < area; i++ {
		realData[offset] = red[i]
		offset++
		realData[offset] = green[i]
		offset++
		realData[offset] = blue[i]
		offset++
		realData[offset] = 255
		offset++
	}

	return &image.RGBA{Pix: realData, Stride: j.width << 2, Rect: j.bounds}
}

func (j *opjp2) decodeRGB16() image.Image {
	var red = jp2ComponentData16(j.comps[0])
	var green = jp2ComponentData16(j.comps[1])
	var blue = jp2ComponentData16(j.comps[2])

	var offset = 0
	var pixels = j.width*j.height
	var realData = make([]uint8, j.width*j.height*8)
	for i := 0; i < pixels; i++ {
		offset = i << 3
		realData[offset] = red[i<<1]
		realData[offset+1] = red[i<<1+1]
		realData[offset+2] = green[i<<1]
		realData[offset+3] = green[i<<1+1]
		realData[offset+4] = blue[i<<1]
		realData[offset+5] = blue[i<<1+1]
		realData[offset+6] = 0xFF
		realData[offset+7] = 0xFF
	}

	return &image.RGBA64{Pix: realData, Stride: j.width << 3, Rect: j.bounds}
}

// jp2ComponentData8 returns a slice of Image-usable uint8s from the JP2 raw
// data in the given openjpeg component
func jp2ComponentData8(comp C.struct_opj_image_comp) []uint8 {
	var data []int32
	dataSlice := (*reflect.SliceHeader)((unsafe.Pointer(&data)))
	size := int(comp.w) * int(comp.h)
	dataSlice.Cap = size
	dataSlice.Len = size
	dataSlice.Data = uintptr(unsafe.Pointer(comp.data))

	realData := make([]uint8, len(data))
	for index, point := range data {
		realData[index] = uint8(point)
	}

	return realData
}

// jp2ComponentData16 returns a slice of Image-usable uint8s for a 16-bit
// grayscale image, using the JP2 raw data in the given openjpeg component
func jp2ComponentData16(comp C.struct_opj_image_comp) []uint8 {
	var data []int32
	dataSlice := (*reflect.SliceHeader)((unsafe.Pointer(&data)))
	size := int(comp.w) * int(comp.h)
	dataSlice.Cap = size
	dataSlice.Len = size
	dataSlice.Data = uintptr(unsafe.Pointer(comp.data))

	realData := make([]uint8, len(data)*2)
	for i, point := range data {
		var val = uint16(point)
		realData[i<<1+0] = uint8(val >> 8)
		realData[i<<1+1] = uint8(val)
	}

	return realData
}
