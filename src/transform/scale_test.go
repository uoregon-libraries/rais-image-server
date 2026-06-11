package transform

import (
	"fmt"
	"image"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

// lcg gives us deterministic pseudo-random pixel data
type lcg uint32

func (l *lcg) next() uint8 {
	*l = *l*1664525 + 1013904223
	return uint8(*l >> 24)
}

func randomImage(kind string, r image.Rectangle) image.Image {
	var seed = lcg(0x1234)
	switch kind {
	case "Gray":
		var i = image.NewGray(r)
		for n := range i.Pix {
			i.Pix[n] = seed.next()
		}
		return i
	case "Gray16":
		var i = image.NewGray16(r)
		for n := range i.Pix {
			i.Pix[n] = seed.next()
		}
		return i
	case "RGBA":
		var i = image.NewRGBA(r)
		for n := range i.Pix {
			i.Pix[n] = seed.next()
		}
		return i
	case "RGBA64":
		var i = image.NewRGBA64(r)
		for n := range i.Pix {
			i.Pix[n] = seed.next()
		}
		return i
	}
	panic("bad image kind " + kind)
}

// refChannels computes a float64 reference bilinear interpolation for the dst
// pixel (x, y), returning the four channels in the 16-bit space At().RGBA()
// uses.  This mirrors the center-aligned sampling Scale is supposed to do.
func refChannels(src image.Image, dstW, dstH, x, y int) [4]float64 {
	var b = src.Bounds()
	var sample = func(sx, sy int) [4]float64 {
		var r, g, bl, a = src.At(b.Min.X+sx, b.Min.Y+sy).RGBA()
		return [4]float64{float64(r), float64(g), float64(bl), float64(a)}
	}
	var coords = func(d, dstDim, srcDim int) (int, int, float64) {
		var c = (float64(d)+0.5)*float64(srcDim)/float64(dstDim) - 0.5
		if c < 0 {
			c = 0
		}
		var i0 = int(c)
		var i1 = i0 + 1
		if i1 > srcDim-1 {
			i1 = srcDim - 1
		}
		return i0, i1, c - float64(i0)
	}

	var x0, x1, fx = coords(x, dstW, b.Dx())
	var y0, y1, fy = coords(y, dstH, b.Dy())
	var p00, p01 = sample(x0, y0), sample(x1, y0)
	var p10, p11 = sample(x0, y1), sample(x1, y1)
	var out [4]float64
	for c := 0; c < 4; c++ {
		var top = p00[c] + (p01[c]-p00[c])*fx
		var bot = p10[c] + (p11[c]-p10[c])*fx
		out[c] = top + (bot-top)*fy
	}
	return out
}

func assertMatchesReference(t *testing.T, src image.Image, dstW, dstH int, tolerance float64) {
	var dst = Scale(src, dstW, dstH)
	if dst == nil {
		t.Fatalf("Scale returned nil for %T", src)
	}
	var b = dst.Bounds()
	assert.Equal(dstW, b.Dx(), "scaled width", t)
	assert.Equal(dstH, b.Dy(), "scaled height", t)

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			var ref = refChannels(src, dstW, dstH, x, y)
			var r, g, bl, a = dst.At(x, y).RGBA()
			var got = [4]float64{float64(r), float64(g), float64(bl), float64(a)}
			for c := 0; c < 4; c++ {
				var diff = got[c] - ref[c]
				if diff < 0 {
					diff = -diff
				}
				if diff > tolerance {
					t.Fatalf("%T scale %dx%d -> %dx%d: pixel (%d, %d) channel %d: got %f, want %f (tolerance %f)",
						src, src.Bounds().Dx(), src.Bounds().Dy(), dstW, dstH, x, y, c, got[c], ref[c], tolerance)
				}
			}
		}
	}
}

func TestScaleMatchesReference(t *testing.T) {
	// 8-bit channels are reported by RGBA() in 16-bit space (one unit is 257),
	// so rounding to the nearest 8-bit value can differ from the float
	// reference by up to half a unit
	var tolerances = map[string]float64{"Gray": 130, "RGBA": 130, "Gray16": 2, "RGBA64": 2}
	var sizes = []struct{ sw, sh, dw, dh int }{
		{13, 7, 29, 17},  // upscale
		{64, 48, 31, 23}, // non-integer downscale
		{32, 32, 16, 16}, // exact 2:1
		{20, 20, 20, 20}, // same size
		{16, 16, 1, 1},   // extreme downscale
	}
	for kind, tolerance := range tolerances {
		for _, s := range sizes {
			t.Run(fmt.Sprintf("%s_%dx%d_to_%dx%d", kind, s.sw, s.sh, s.dw, s.dh), func(t *testing.T) {
				assertMatchesReference(t, randomImage(kind, image.Rect(0, 0, s.sw, s.sh)), s.dw, s.dh, tolerance)
			})
		}
	}
}

// TestScaleSubImage ensures sources with non-zero bounds (e.g., from
// SubImage) are read from the right place
func TestScaleSubImage(t *testing.T) {
	var full = randomImage("Gray", image.Rect(0, 0, 64, 64)).(*image.Gray)
	var sub = full.SubImage(image.Rect(16, 8, 48, 40)).(*image.Gray)
	assertMatchesReference(t, sub, 16, 16, 130)

	var full16 = randomImage("RGBA64", image.Rect(0, 0, 64, 64)).(*image.RGBA64)
	var sub16 = full16.SubImage(image.Rect(16, 8, 48, 40)).(*image.RGBA64)
	assertMatchesReference(t, sub16, 16, 16, 2)
}

// TestScaleConstant verifies a solid-color image stays exactly solid - any
// deviation means our weights don't sum to one
func TestScaleConstant(t *testing.T) {
	var src = image.NewGray16(image.Rect(0, 0, 50, 30))
	for n := 0; n < len(src.Pix); n += 2 {
		src.Pix[n], src.Pix[n+1] = 0xBE, 0xEF
	}
	var dst = Scale(src, 33, 21).(*image.Gray16)
	for n := 0; n < len(dst.Pix); n += 2 {
		if dst.Pix[n] != 0xBE || dst.Pix[n+1] != 0xEF {
			t.Fatalf("constant image changed at pix offset %d: %#x%x", n, dst.Pix[n], dst.Pix[n+1])
		}
	}
}

// TestScaleCheckerboard verifies an exact 2:1 downscale of a checkerboard
// averages each 2x2 block (the half-pixel weights make this case exact)
func TestScaleCheckerboard(t *testing.T) {
	var src = image.NewGray(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			if (x+y)%2 == 0 {
				src.Pix[y*src.Stride+x] = 255
			}
		}
	}
	var dst = Scale(src, 16, 16).(*image.Gray)
	for n, p := range dst.Pix {
		// 127.5 rounds up
		if p != 128 {
			t.Fatalf("checkerboard downscale: expected 128 at pix %d, got %d", n, p)
		}
	}
}

func TestScaleUnsupportedType(t *testing.T) {
	var dst = Scale(image.NewNRGBA(image.Rect(0, 0, 8, 8)), 4, 4)
	if dst != nil {
		t.Fatalf("expected nil for unsupported image type, got %T", dst)
	}
}

func benchSetup(kind string) image.Image {
	return randomImage(kind, image.Rect(0, 0, 2048, 2048))
}

func BenchmarkScaleGray(b *testing.B) {
	var src = benchSetup("Gray")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Scale(src, 1024, 1024)
	}
}

func BenchmarkScaleRGBA(b *testing.B) {
	var src = benchSetup("RGBA")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Scale(src, 1024, 1024)
	}
}

func BenchmarkScaleGray16(b *testing.B) {
	var src = benchSetup("Gray16")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Scale(src, 1024, 1024)
	}
}

func BenchmarkScaleRGBA64(b *testing.B) {
	var src = benchSetup("RGBA64")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		Scale(src, 1024, 1024)
	}
}
