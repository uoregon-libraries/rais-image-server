package openjpeg

import (
	"image"
	"math"
)

const MAX_PROGRESSION_LEVEL = uint(6)

func min(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

// Returns the scale in powers of two between two numbers
func getScale(v1, v2 int) uint {
	if v1 == v2 {
		return 0
	}

	large, small := float64(v1), float64(v2)
	if large < small {
		large, small = small, large
	}

	return uint(math.Floor(math.Log2(large) - math.Log2(small)))
}

func desired_progression_level(r image.Rectangle, width, height int) uint {
	if width > r.Dx() || height > r.Dy() {
		return 0
	}

	// If either dimension is zero, we want to avoid computation and just use the
	// other's scale value
	scaleX := MAX_PROGRESSION_LEVEL
	scaleY := MAX_PROGRESSION_LEVEL

	if width > 0 {
		scaleX = getScale(r.Dx(), width)
	}

	if height > 0 {
		scaleY = getScale(r.Dy(), height)
	}

	// Pull the smallest value - if we request a resize from 1000x1000 to 250x500
	// (for some odd reason), then we need to start with the 500x500 level, not
	// the 250x250 level
	level := min(scaleX, scaleY)
	return min(MAX_PROGRESSION_LEVEL, level)
}
