// GENERATED CODE; DO NOT EDIT!

package transform

import (
	"image"
)

type Rotator interface {
	Image() image.Image
	Rotate90()
	Rotate180()
	Rotate270()
	Mirror()
}

// GrayRotator decorates *image.Gray with rotation functions
type GrayRotator struct {
	Img *image.Gray
}

// Image returns the underlying image as an image.Image value
func (r *GrayRotator) Image() image.Image {
	return r.Img
}

// Rotate90 does a simple 90-degree clockwise rotation
func (r *GrayRotator) Rotate90() {
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
}

// Rotate180 does a simple 180-degree clockwise rotation
func (r *GrayRotator) Rotate180() {
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
}

// Rotate270 does a simple 270-degree clockwise rotation
func (r *GrayRotator) Rotate270() {
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
}

// Mirror flips the image around its vertical axis
func (r *GrayRotator) Mirror() {
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
}

// RGBARotator decorates *image.RGBA with rotation functions
type RGBARotator struct {
	Img *image.RGBA
}

// Image returns the underlying image as an image.Image value
func (r *RGBARotator) Image() image.Image {
	return r.Img
}

// Rotate90 does a simple 90-degree clockwise rotation
func (r *RGBARotator) Rotate90() {
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
}

// Rotate180 does a simple 180-degree clockwise rotation
func (r *RGBARotator) Rotate180() {
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
}

// Rotate270 does a simple 270-degree clockwise rotation
func (r *RGBARotator) Rotate270() {
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
}

// Mirror flips the image around its vertical axis
func (r *RGBARotator) Mirror() {
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
}

// GENERATED CODE; DO NOT EDIT!
