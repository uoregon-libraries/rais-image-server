package iiif

import (
	"fmt"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/color-assert"
	"strings"
	"testing"
)

var weirdID = "identifier-foo-bar%2Fbaz,,,,,chameleon"
var simplePath = "/images/iiif/" + weirdID + "/full/full/30/default.jpg"

func TestInvalid(t *testing.T) {
	badRegion := strings.Replace(simplePath, "/full/full", "/bad/full", 1)
	assert.True(!NewIIIFCommand(badRegion).Valid(), "Expected bad region string to be invalid", t)
}

func TestValid(t *testing.T) {
	i := NewIIIFCommand(simplePath)

	assert.True(i.Valid(), fmt.Sprintf("Expected %s to be valid", simplePath), t)
	assert.Equal(weirdID, i.ID, "identifier should be extracted", t)
	assert.Equal(RTFull, i.Region.Type, "Region is RTFull", t)
	assert.Equal(STFull, i.Size.Type, "Size is STFull", t)
	assert.Equal(30.0, i.Rotation.Degrees, "i.Rotation.Degrees", t)
	assert.True(!i.Rotation.Mirror, "!i.Rotation.Mirror", t)
	assert.Equal(QDefault, i.Quality, "i.Quality == QDefault", t)
	assert.Equal(FmtJPG, i.Format, "i.Format == FmtJPG", t)
}
