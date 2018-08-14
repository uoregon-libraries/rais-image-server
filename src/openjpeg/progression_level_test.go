package openjpeg

import (
	"image"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

func TestDesiredProgressionLevel(t *testing.T) {
	source := image.Rect(0, 0, 5000, 5000)
	dpl := func(w, h int) int {
		return desiredProgressionLevel(source, w, h)
	}

	assert.Equal(0, dpl(5000, 5000), "Source and dest are equal, level should be 0", t)
	assert.Equal(0, dpl(5001, 5001), "Source is SMALLER than dest, so level has to be 0", t)
	assert.Equal(0, dpl(4999, 4999), "Source is larger than dest, but not by a factor of 2, so level has to be 0", t)
	assert.Equal(0, dpl(2501, 2501), "Source is just under 2x dest, so level still has to be 0", t)
	assert.Equal(1, dpl(2500, 2500), "Source is exactly 2x dest, so level has to be 1", t)
	assert.Equal(1, dpl(2500, 250), "We have to pick the largest dimension, so level for 2500x250 should be 1", t)
	assert.Equal(2, dpl(1250, 0), "Use the non-zero dimension for resize package compatibility", t)
}
