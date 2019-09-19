package iiif

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

var weirdID = "identifier-foo-bar/baz,,,,,chameleon"
var simplePath = url.QueryEscape(weirdID) + "/full/full/30/default.jpg"

func TestInvalid(t *testing.T) {
	badURL := strings.Replace(simplePath, "/full/full", "/bad/full", 1)
	badURL = strings.Replace(badURL, "default.jpg", "default.foo", 1)
	i, err := NewURL(badURL)
	assert.Equal("invalid region, invalid format", err.Error(), "NewURL error message", t)
	assert.False(i.Valid(), "IIIF URL is invalid", t)

	// All other data should still be extracted despite this being a bad IIIF URL
	assert.Equal(weirdID, string(i.ID), "identifier should be extracted", t)
	assert.Equal(RTNone, i.Region.Type, "bad Region is RTNone", t)
	assert.Equal(STFull, i.Size.Type, "Size is STFull", t)
	assert.Equal(30.0, i.Rotation.Degrees, "i.Rotation.Degrees", t)
	assert.True(!i.Rotation.Mirror, "!i.Rotation.Mirror", t)
	assert.Equal(QDefault, i.Quality, "i.Quality == QDefault", t)
	assert.Equal(FmtUnknown, i.Format, "i.Format == FmtJPG", t)
	assert.Equal(false, i.Info, "not an info request", t)
}

func TestValid(t *testing.T) {
	i, err := NewURL(simplePath)
	assert.NilError(err, "NewURL has no error", t)

	assert.True(i.Valid(), fmt.Sprintf("Expected %s to be valid", simplePath), t)
	assert.Equal(weirdID, string(i.ID), "identifier should be extracted", t)
	assert.Equal(RTFull, i.Region.Type, "Region is RTFull", t)
	assert.Equal(STFull, i.Size.Type, "Size is STFull", t)
	assert.Equal(30.0, i.Rotation.Degrees, "i.Rotation.Degrees", t)
	assert.True(!i.Rotation.Mirror, "!i.Rotation.Mirror", t)
	assert.Equal(QDefault, i.Quality, "i.Quality == QDefault", t)
	assert.Equal(FmtJPG, i.Format, "i.Format == FmtJPG", t)
	assert.Equal(false, i.Info, "not an info request", t)
}

func TestInfo(t *testing.T) {
	i, err := NewURL("some%2Fvalid%2Fpath.jp2/info.json")
	assert.NilError(err, "info request isn't an error", t)
	assert.Equal("some/valid/path.jp2", string(i.ID), "identifier", t)
	assert.Equal(true, i.Info, "is an info request", t)
}

func TestInfoBaseRedirect(t *testing.T) {
	i, err := NewURL("some%2Fvalid%2Fpath.jp2")
	assert.Equal("empty id, invalid region, invalid size, invalid quality", err.Error(), "base redirects are error cases the caller must handle", t)
	assert.Equal("", string(i.ID), "identifier", t)
}
