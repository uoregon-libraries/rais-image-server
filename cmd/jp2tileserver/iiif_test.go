package main

import (
	"fmt"
	"strings"
	"testing"
)

var weirdID = "identifier-foo-bar%2Fbaz,,,,,chameleon"
var simplePath = "/images/iiif/" + weirdID + "/full/full/30/default.jpg"

func TestInvalid(t *testing.T) {
	badRegion := strings.Replace(simplePath, "/full/full", "/bad/full", 1)
	assert(!NewIIIFCommand(badRegion).Valid(), "Expected bad region string to be invalid", t)
}

func TestValid(t *testing.T) {
	i := NewIIIFCommand(simplePath)

	assert(i.Valid(), fmt.Sprintf("Expected %s to be valid", simplePath), t)
	assertEqual(weirdID, i.ID, "identifier should be extracted", t)
	assertEqual(RTFull, i.Region.Type, "Region is RTFull", t)
	assertEqual(STFull, i.Size.Type, "Size is STFull", t)
	assertEqual(30.0, i.Rotation.Degrees, "i.Rotation.Degrees", t)
	assert(!i.Rotation.Mirror, "!i.Rotation.Mirror", t)
	assertEqual(QDefault, i.Quality, "i.Quality == QDefault", t)
	assertEqual(FmtJPG, i.Format, "i.Format == FmtJPG", t)
}
