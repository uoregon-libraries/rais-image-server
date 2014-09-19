package openjpeg

import (
	"image"
	"image/color"
)

type RawImage struct {
	data   []int32
	bounds image.Rectangle
	stride int
}

func (p *RawImage) ColorModel() color.Model {
	return color.GrayModel
}

func (p *RawImage) Bounds() image.Rectangle {
	return p.bounds
}

func (p *RawImage) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.bounds)) {
		return color.Gray{}
	}
	index := p.PixOffset(x, y)
	return color.Gray{uint8(p.data[index])}
}

func (p *RawImage) PixOffset(x, y int) int {
	return (y-p.bounds.Min.Y)*p.stride + (x-p.bounds.Min.X)*1
}
