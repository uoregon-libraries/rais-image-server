package openjpeg

import (
	"image"
	"testing"
)

func assertEqualUInt(expected, actual uint, message string, t *testing.T) {
	if expected != actual {
		t.Errorf("Expected %d, but got %d - %s", expected, actual, message)
		return
	}
	t.Log(message)
}

func TestDesiredProgressionLevel(t *testing.T) {
	source := image.Rect(0, 0, 5000, 5000)
	dpl := func(w, h int) uint {
		return desired_progression_level(source, w, h)
	}

	assertEqualUInt(0, dpl(5000, 5000), "Source and dest are equal, level should be 0", t)
	assertEqualUInt(0, dpl(5001, 5001), "Source is SMALLER than dest, so level has to be 0", t)
	assertEqualUInt(0, dpl(4999, 4999), "Source is larger than dest, but not by a factor of 2, so level has to be 0", t)
	assertEqualUInt(0, dpl(2501, 2501), "Source is just under 2x dest, so level still has to be 0", t)
	assertEqualUInt(1, dpl(2500, 2500), "Source is exactly 2x dest, so level has to be 1", t)
	assertEqualUInt(1, dpl(2500, 250), "We have to pick the largest dimension, so level for 2500x250 should be 1", t)
	assertEqualUInt(2, dpl(1250, 0), "Use the non-zero dimension for resize package compatibility", t)
}
