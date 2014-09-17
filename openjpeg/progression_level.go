package openjpeg

import (
	"image"
)

const MAX_PROGRESSION_LEVEL = uint(6)

func scaled_dimension(progression_level uint, dimension int) int {
	scale_factor := uint(2) << (progression_level - uint(1))
	return int(float32(dimension) / float32(scale_factor))
}

func desired_progression_level(r image.Rectangle, width, height int) uint {
	level := MAX_PROGRESSION_LEVEL
	for ; level > 0 && width > scaled_dimension(level, r.Dx()) && height > scaled_dimension(level, r.Dy()); level-- {
	}
	return level
}

