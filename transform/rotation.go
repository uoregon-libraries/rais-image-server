// GENERATED CODE; DO NOT EDIT!

package transform

import (
	"image"
)

// Rotator implements simple 90-degree rotations in addition to mirroring for
// IIIF compliance.  After each operation, the underlying image is replaced
// with the new image.  It's important to note, however, that the source image
// is never directly changed.  A new image is drawn, and the old is simply
// forgotten by the Rotator.
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

	var x, y, srcIdx, dstIdx int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	srcPix := src.Pix
	dstPix := dst.Pix
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcIdx = y*srcStride + x
			dstIdx = x*dstStride + (maxY - 1 - y)
			dstPix[dstIdx] = srcPix[srcIdx]
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

	var x, y, srcIdx, dstIdx int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	srcPix := src.Pix
	dstPix := dst.Pix
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcIdx = y*srcStride + x
			dstIdx = (maxY-1-y)*dstStride + (maxX - 1 - x)
			dstPix[dstIdx] = srcPix[srcIdx]
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

	var x, y, srcIdx, dstIdx int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	srcPix := src.Pix
	dstPix := dst.Pix
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcIdx = y*srcStride + x
			dstIdx = (maxX-1-x)*dstStride + y
			dstPix[dstIdx] = srcPix[srcIdx]
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

	var x, y, srcIdx, dstIdx int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	srcPix := src.Pix
	dstPix := dst.Pix
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcIdx = y*srcStride + x
			dstIdx = y*dstStride + (maxX - 1 - x)
			dstPix[dstIdx] = srcPix[srcIdx]
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

	var x, y, srcIdx, dstIdx int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	srcPix := src.Pix
	dstPix := dst.Pix
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcIdx = y*srcStride + (x << 2)
			dstIdx = x*dstStride + ((maxY - 1 - y) << 2)
			copy(dstPix[dstIdx:dstIdx+4], srcPix[srcIdx:srcIdx+4])
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

	var x, y, srcIdx, dstIdx int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	srcPix := src.Pix
	dstPix := dst.Pix
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcIdx = y*srcStride + (x << 2)
			dstIdx = (maxY-1-y)*dstStride + ((maxX - 1 - x) << 2)
			copy(dstPix[dstIdx:dstIdx+4], srcPix[srcIdx:srcIdx+4])
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

	var x, y, srcIdx, dstIdx int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	srcPix := src.Pix
	dstPix := dst.Pix
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcIdx = y*srcStride + (x << 2)
			dstIdx = (maxX-1-x)*dstStride + (y << 2)
			copy(dstPix[dstIdx:dstIdx+4], srcPix[srcIdx:srcIdx+4])
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

	var x, y, srcIdx, dstIdx int64
	maxX, maxY := int64(srcWidth), int64(srcHeight)
	srcStride, dstStride := int64(src.Stride), int64(dst.Stride)
	srcPix := src.Pix
	dstPix := dst.Pix
	for y = 0; y < maxY; y++ {
		for x = 0; x < maxX; x++ {
			srcIdx = y*srcStride + (x << 2)
			dstIdx = y*dstStride + ((maxX - 1 - x) << 2)
			copy(dstPix[dstIdx:dstIdx+4], srcPix[srcIdx:srcIdx+4])
		}
	}

	r.Img = dst
}

// GENERATED CODE; DO NOT EDIT!
