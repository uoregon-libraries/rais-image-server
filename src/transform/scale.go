package transform

import (
	"image"
)

// Scale resizes src to dstW x dstH using bilinear interpolation, returning
// the scaled image.  Only the four concrete image types RAIS decoders produce
// are supported (Gray, Gray16, RGBA, and RGBA64); anything else returns nil.
//
// These hand-rolled scalers exist because golang.org/x/image/draw only has
// fast paths for *image.RGBA destinations: scaling grayscale or 16-bit images
// with it goes through a generic per-pixel path that's over an order of
// magnitude slower than this implementation.
func Scale(src image.Image, dstW, dstH int) image.Image {
	switch s := src.(type) {
	case *image.Gray:
		return ScaleGray(s, dstW, dstH)
	case *image.Gray16:
		return ScaleGray16(s, dstW, dstH)
	case *image.RGBA:
		return ScaleRGBA(s, dstW, dstH)
	case *image.RGBA64:
		return ScaleRGBA64(s, dstW, dstH)
	}
	return nil
}

// bilinearCoords precomputes, for each destination index along one axis, the
// two source indexes to sample (relative to the source bounds) and the 16.16
// fixed-point weight of the second sample.  Sample positions are
// center-aligned: source position = (dst + 0.5) * srcDim/dstDim - 0.5.
func bilinearCoords(srcDim, dstDim int) (i0, i1, frac []int) {
	i0 = make([]int, dstDim)
	i1 = make([]int, dstDim)
	frac = make([]int, dstDim)
	for x := 0; x < dstDim; x++ {
		var c = (int64(2*x+1)*int64(srcDim)<<15)/int64(dstDim) - 1<<15
		if c < 0 {
			c = 0
		}
		var x0 = int(c >> 16)
		var x1 = x0 + 1
		if x1 > srcDim-1 {
			x1 = srcDim - 1
		}
		i0[x], i1[x], frac[x] = x0, x1, int(c&0xffff)
	}
	return i0, i1, frac
}

// bilerp8 interpolates four 8-bit samples with 16.16 fixed-point x and y
// weights, rounding to nearest
func bilerp8(p00, p01, p10, p11 uint8, fx, fy int64) uint8 {
	var top = int64(p00)<<16 + (int64(p01)-int64(p00))*fx
	var bot = int64(p10)<<16 + (int64(p11)-int64(p10))*fx
	return uint8((top<<16 + (bot-top)*fy + 1<<31) >> 32)
}

// bilerp16 interpolates four 16-bit samples with 16.16 fixed-point x and y
// weights, rounding to nearest
func bilerp16(p00, p01, p10, p11 uint16, fx, fy int64) uint16 {
	var top = int64(p00)<<16 + (int64(p01)-int64(p00))*fx
	var bot = int64(p10)<<16 + (int64(p11)-int64(p10))*fx
	return uint16((top<<16 + (bot-top)*fy + 1<<31) >> 32)
}

// be16 reads a big-endian uint16 from pix at offset o
func be16(pix []uint8, o int) uint16 {
	return uint16(pix[o])<<8 | uint16(pix[o+1])
}

// putbe16 writes v to pix at offset o in big-endian order
func putbe16(pix []uint8, o int, v uint16) {
	pix[o] = uint8(v >> 8)
	pix[o+1] = uint8(v)
}

// ScaleGray resizes src to dstW x dstH using bilinear interpolation
func ScaleGray(src *image.Gray, dstW, dstH int) *image.Gray {
	var dst = image.NewGray(image.Rect(0, 0, dstW, dstH))
	var b = src.Bounds()
	var x0s, x1s, xfs = bilinearCoords(b.Dx(), dstW)
	var y0s, y1s, yfs = bilinearCoords(b.Dy(), dstH)
	var base = src.PixOffset(b.Min.X, b.Min.Y)

	for y := 0; y < dstH; y++ {
		var row0 = src.Pix[base+y0s[y]*src.Stride:]
		var row1 = src.Pix[base+y1s[y]*src.Stride:]
		var fy = int64(yfs[y])
		var drow = dst.Pix[y*dst.Stride : y*dst.Stride+dstW]
		for x := 0; x < dstW; x++ {
			var x0, x1, fx = x0s[x], x1s[x], int64(xfs[x])
			drow[x] = bilerp8(row0[x0], row0[x1], row1[x0], row1[x1], fx, fy)
		}
	}

	return dst
}

// ScaleGray16 resizes src to dstW x dstH using bilinear interpolation
func ScaleGray16(src *image.Gray16, dstW, dstH int) *image.Gray16 {
	var dst = image.NewGray16(image.Rect(0, 0, dstW, dstH))
	var b = src.Bounds()
	var x0s, x1s, xfs = bilinearCoords(b.Dx(), dstW)
	var y0s, y1s, yfs = bilinearCoords(b.Dy(), dstH)
	var base = src.PixOffset(b.Min.X, b.Min.Y)

	for y := 0; y < dstH; y++ {
		var row0 = src.Pix[base+y0s[y]*src.Stride:]
		var row1 = src.Pix[base+y1s[y]*src.Stride:]
		var fy = int64(yfs[y])
		var drow = dst.Pix[y*dst.Stride:]
		for x := 0; x < dstW; x++ {
			var s0, s1, fx = x0s[x] << 1, x1s[x] << 1, int64(xfs[x])
			putbe16(drow, x<<1, bilerp16(be16(row0, s0), be16(row0, s1), be16(row1, s0), be16(row1, s1), fx, fy))
		}
	}

	return dst
}

// ScaleRGBA resizes src to dstW x dstH using bilinear interpolation.  All
// four channels are interpolated independently, which is correct for the
// premultiplied alpha RGBA uses.
func ScaleRGBA(src *image.RGBA, dstW, dstH int) *image.RGBA {
	var dst = image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	var b = src.Bounds()
	var x0s, x1s, xfs = bilinearCoords(b.Dx(), dstW)
	var y0s, y1s, yfs = bilinearCoords(b.Dy(), dstH)
	var base = src.PixOffset(b.Min.X, b.Min.Y)

	for y := 0; y < dstH; y++ {
		var row0 = src.Pix[base+y0s[y]*src.Stride:]
		var row1 = src.Pix[base+y1s[y]*src.Stride:]
		var fy = int64(yfs[y])
		var drow = dst.Pix[y*dst.Stride:]
		for x := 0; x < dstW; x++ {
			var s0, s1, fx = x0s[x] << 2, x1s[x] << 2, int64(xfs[x])
			var o = x << 2
			drow[o] = bilerp8(row0[s0], row0[s1], row1[s0], row1[s1], fx, fy)
			drow[o+1] = bilerp8(row0[s0+1], row0[s1+1], row1[s0+1], row1[s1+1], fx, fy)
			drow[o+2] = bilerp8(row0[s0+2], row0[s1+2], row1[s0+2], row1[s1+2], fx, fy)
			drow[o+3] = bilerp8(row0[s0+3], row0[s1+3], row1[s0+3], row1[s1+3], fx, fy)
		}
	}

	return dst
}

// ScaleRGBA64 resizes src to dstW x dstH using bilinear interpolation.  All
// four channels are interpolated independently, which is correct for the
// premultiplied alpha RGBA64 uses.
func ScaleRGBA64(src *image.RGBA64, dstW, dstH int) *image.RGBA64 {
	var dst = image.NewRGBA64(image.Rect(0, 0, dstW, dstH))
	var b = src.Bounds()
	var x0s, x1s, xfs = bilinearCoords(b.Dx(), dstW)
	var y0s, y1s, yfs = bilinearCoords(b.Dy(), dstH)
	var base = src.PixOffset(b.Min.X, b.Min.Y)

	for y := 0; y < dstH; y++ {
		var row0 = src.Pix[base+y0s[y]*src.Stride:]
		var row1 = src.Pix[base+y1s[y]*src.Stride:]
		var fy = int64(yfs[y])
		var drow = dst.Pix[y*dst.Stride:]
		for x := 0; x < dstW; x++ {
			var s0, s1, fx = x0s[x] << 3, x1s[x] << 3, int64(xfs[x])
			var o = x << 3
			for c := 0; c < 8; c += 2 {
				putbe16(drow, o+c, bilerp16(be16(row0, s0+c), be16(row0, s1+c), be16(row1, s0+c), be16(row1, s1+c), fx, fy))
			}
		}
	}

	return dst
}
