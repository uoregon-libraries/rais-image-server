// GENERATED CODE; DO NOT EDIT!

package transform

import (
	"image"
)

type Rotator interface {
	Rotate90() image.Image
	Rotate180() image.Image
	Rotate270() image.Image
	Mirror() image.Image
}

// GrayRotator decorates *image.Gray with rotation functions
type GrayRotator struct {
	Img *image.Gray
}

// Rotate90 does a simple 90-degree clockwise rotation, returning a new image.Image
func (r *GrayRotator) Rotate90() image.Image {
	src := r.Img
	srcB := src.Bounds()
	srcWidth := srcB.Dx()
	srcHeight := srcB.Dy()

	dst := image.NewGray(image.Rect(0, 0, srcHeight, srcWidth))

	var x, y, srcPix, dstPix int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcPix = y*srcStride + x
			dstPix = x*dstStride + (maxY - 1 - y)
			dst.Pix[dstPix] = src.Pix[srcPix]
		}
	}

	r.Img = dst
	return dst
}

// Rotate180 does a simple 180-degree clockwise rotation, returning a new image.Image
func (r *GrayRotator) Rotate180() image.Image {
	src := r.Img
	srcB := src.Bounds()
	srcWidth := srcB.Dx()
	srcHeight := srcB.Dy()

	dst := image.NewGray(image.Rect(0, 0, srcWidth, srcHeight))

	var x, y, srcPix, dstPix int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcPix = y*srcStride + x
			dstPix = (maxY-1-y)*dstStride + (maxX - 1 - x)
			dst.Pix[dstPix] = src.Pix[srcPix]
		}
	}

	r.Img = dst
	return dst
}

// Rotate270 does a simple 270-degree clockwise rotation, returning a new image.Image
func (r *GrayRotator) Rotate270() image.Image {
	src := r.Img
	srcB := src.Bounds()
	srcWidth := srcB.Dx()
	srcHeight := srcB.Dy()

	dst := image.NewGray(image.Rect(0, 0, srcHeight, srcWidth))

	var x, y, srcPix, dstPix int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcPix = y*srcStride + x
			dstPix = (maxX-1-x)*dstStride + y
			dst.Pix[dstPix] = src.Pix[srcPix]
		}
	}

	r.Img = dst
	return dst
}

// Mirror flips the image around its vertical axis, returning a new image.Image
func (r *GrayRotator) Mirror() image.Image {
	src := r.Img
	srcB := src.Bounds()
	srcWidth := srcB.Dx()
	srcHeight := srcB.Dy()

	dst := image.NewGray(image.Rect(0, 0, srcWidth, srcHeight))

	var x, y, srcPix, dstPix int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcPix = y*srcStride + x
			dstPix = y*dstStride + (maxX - 1 - x)
			dst.Pix[dstPix] = src.Pix[srcPix]
		}
	}

	r.Img = dst
	return dst
}

// RGBARotator decorates *image.RGBA with rotation functions
type RGBARotator struct {
	Img *image.RGBA
}

// Rotate90 does a simple 90-degree clockwise rotation, returning a new image.Image
func (r *RGBARotator) Rotate90() image.Image {
	src := r.Img
	srcB := src.Bounds()
	srcWidth := srcB.Dx()
	srcHeight := srcB.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, srcHeight, srcWidth))

	var x, y, srcPix, dstPix int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcPix = y*srcStride + (x << 2)
			dstPix = x*dstStride + ((maxY - 1 - y) << 2)
			copy(dst.Pix[dstPix:dstPix+4], src.Pix[srcPix:srcPix+4])
		}
	}

	r.Img = dst
	return dst
}

// Rotate180 does a simple 180-degree clockwise rotation, returning a new image.Image
func (r *RGBARotator) Rotate180() image.Image {
	src := r.Img
	srcB := src.Bounds()
	srcWidth := srcB.Dx()
	srcHeight := srcB.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, srcWidth, srcHeight))

	var x, y, srcPix, dstPix int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcPix = y*srcStride + (x << 2)
			dstPix = (maxY-1-y)*dstStride + ((maxX - 1 - x) << 2)
			copy(dst.Pix[dstPix:dstPix+4], src.Pix[srcPix:srcPix+4])
		}
	}

	r.Img = dst
	return dst
}

// Rotate270 does a simple 270-degree clockwise rotation, returning a new image.Image
func (r *RGBARotator) Rotate270() image.Image {
	src := r.Img
	srcB := src.Bounds()
	srcWidth := srcB.Dx()
	srcHeight := srcB.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, srcHeight, srcWidth))

	var x, y, srcPix, dstPix int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcPix = y*srcStride + (x << 2)
			dstPix = (maxX-1-x)*dstStride + (y << 2)
			copy(dst.Pix[dstPix:dstPix+4], src.Pix[srcPix:srcPix+4])
		}
	}

	r.Img = dst
	return dst
}

// Mirror flips the image around its vertical axis, returning a new image.Image
func (r *RGBARotator) Mirror() image.Image {
	src := r.Img
	srcB := src.Bounds()
	srcWidth := srcB.Dx()
	srcHeight := srcB.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, srcWidth, srcHeight))

	var x, y, srcPix, dstPix int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcPix = y*srcStride + (x << 2)
			dstPix = y*dstStride + ((maxX - 1 - x) << 2)
			copy(dst.Pix[dstPix:dstPix+4], src.Pix[srcPix:srcPix+4])
		}
	}

	r.Img = dst
	return dst
}

// GENERATED CODE; DO NOT EDIT!
