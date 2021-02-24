package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"rais/src/fakehttp"
	"rais/src/iiif"
	"rais/src/img"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/uoregon-libraries/gopkg/assert"
	"github.com/uoregon-libraries/gopkg/logger"
)

func nc(w, h int, a int64) img.Constraint {
	return img.Constraint{Width: w, Height: h, Area: a}
}

var unlimited = nc(math.MaxInt32, math.MaxInt32, math.MaxInt64)

func init() {
	Logger = logger.New(logger.Warn)
	img.RegisterDecodeHandler(decodeJP2)
	img.RegisterStreamReader(fileStreamReader)
}

func rootDir() string {
	p, _ := os.Getwd()
	root, _ := filepath.Abs(p + "/../../../")
	return root
}

// Sets up everything necessary to test a IIIF request
func dorequestGeneric(path string, acceptLD bool, max img.Constraint, fs *iiif.FeatureSet, t *testing.T) *fakehttp.ResponseWriter {
	u, _ := url.Parse("http://example.com")
	w := fakehttp.NewResponseWriter()
	reqPath := fmt.Sprintf("/foo/bar/%s", path)
	req, err := http.NewRequest("get", reqPath, strings.NewReader(""))
	req.RequestURI = reqPath

	if acceptLD {
		req.Header.Add("Accept", "application/ld+json")
	}

	if err != nil {
		t.Errorf("Unable to create fake request: %s", err)
	}
	h := NewImageHandler(rootDir(), "/foo/bar")
	h.Maximums.Width = max.Width
	h.Maximums.Height = max.Height
	h.Maximums.Area = max.Area
	h.BaseURL = u
	h.FeatureSet = fs
	h.IIIFRoute(w, req)

	return w
}

// Sets up everything necessary to test a IIIF request using level 1 support
func dorequest(path string, acceptLD bool, max img.Constraint, t *testing.T) *fakehttp.ResponseWriter {
	return dorequestGeneric(path, acceptLD, max, iiif.FeatureSet1(), t)
}

// Sets up everything necessary to test a IIIF request using level 2 support
func dorequestl2(path string, acceptLD bool, max img.Constraint, t *testing.T) *fakehttp.ResponseWriter {
	return dorequestGeneric(path, acceptLD, max, iiif.FeatureSet2(), t)
}

func request(path string, t *testing.T) *fakehttp.ResponseWriter {
	return dorequest(path, false, unlimited, t)
}

func requestLD(path string, t *testing.T) *fakehttp.ResponseWriter {
	return dorequest(path, true, unlimited, t)
}

func TestInfoHandler404(t *testing.T) {
	w := request("identifier/info.json", t)
	assert.Equal(404, w.StatusCode, "Invalid info request returns 404", t)
}

func TestInfoHandlerJSONOverride(t *testing.T) {
	w := request("docker%2Fimages%2Ftestfile%2Ftest-world.jp2/info.json", t)
	assert.Equal(-1, w.StatusCode, "Valid info request doesn't explicitly set status code", t)
	var data iiif.Info
	json.Unmarshal(w.Output, &data)
	assert.Equal("http://iiif.io/api/image/2/level2.json", data.Profile.ConformanceURL, "Proper profile string", t)
	assert.Equal(800, data.Width, "JSON-decoded width", t)
	assert.Equal(400, data.Height, "JSON-decoded height", t)
	assert.Equal(512, data.Tiles[0].Width, "JSON-decoded tile width", t)
	assert.Equal(0, data.Tiles[0].Height, "JSON-decoded tile height", t)
	assert.Equal(2, len(data.Tiles[0].ScaleFactors), "1 scale factor exists", t)
	assert.Equal(1, data.Tiles[0].ScaleFactors[0], "First scale factor is 1", t)
	assert.Equal(2, data.Tiles[0].ScaleFactors[1], "Second scale factor is 1", t)
	assert.Equal("http://example.com/foo/bar/docker%2Fimages%2Ftestfile%2Ftest-world.jp2", data.ID, "JSON-decoded ID", t)
	assert.Equal(1, len(w.Headers["Content-Type"]), "Proper content type length", t)
	assert.Equal("application/json", w.Headers["Content-Type"][0], "Proper content type", t)
}

func TestInfoHandlerBuiltJSON(t *testing.T) {
	// We don't want to test the JSON override this time, so we use the symlink
	w := request("docker%2Fimages%2Ftestfile%2Ftest-world-link.jp2/info.json", t)
	assert.Equal(-1, w.StatusCode, "Valid info request doesn't explicitly set status code", t)
	var data iiif.Info
	json.Unmarshal(w.Output, &data)
	assert.Equal("http://iiif.io/api/image/2/level1.json", data.Profile.ConformanceURL, "Proper profile string", t)
	assert.Equal(800, data.Width, "JSON-decoded width", t)
	assert.Equal(400, data.Height, "JSON-decoded height", t)
	assert.Equal(0, len(data.Tiles), "Tiles aren't reported when full image is a single tile", t)
	assert.Equal("http://example.com/foo/bar/docker%2Fimages%2Ftestfile%2Ftest-world-link.jp2", data.ID, "JSON-decoded ID", t)
	assert.Equal(1, len(w.Headers["Content-Type"]), "Proper content type length", t)
	assert.Equal("application/json", w.Headers["Content-Type"][0], "Proper content type", t)
}

func TestInfoHandlerLD(t *testing.T) {
	w := requestLD("docker%2Fimages%2Ftestfile%2Ftest-world.jp2/info.json", t)
	assert.Equal(-1, w.StatusCode, "Valid info request doesn't explicitly set status code", t)
	assert.Equal(1, len(w.Headers["Content-Type"]), "Proper content type length", t)
	assert.Equal("application/ld+json", w.Headers["Content-Type"][0], "Proper content type", t)
}

func TestInfoRedirect(t *testing.T) {
	w := request("docker%2Fimages%2Ftestfile%2Ftest-world.jp2", t)
	assert.Equal(303, w.StatusCode, "Base URL redirects to info request", t)
	locHeader := w.Headers["Location"]
	assert.Equal(1, len(locHeader), "There's only 1 location header", t)
	assert.Equal("/foo/bar/docker%2Fimages%2Ftestfile%2Ftest-world.jp2/info.json", locHeader[0], "Location header", t)
}

// TestInfoMaxSize verifies that when the image is bigger than the handler's
// maximums, values are present in the info profile
func TestInfoMaxSize(t *testing.T) {
	w := dorequest("docker%2Fimages%2Ftestfile%2Ftest-world-link.jp2/info.json", false, nc(60, 80, 480), t)
	var data iiif.Info
	json.Unmarshal(w.Output, &data)
	assert.Equal(60, data.Profile.MaxWidth, "JSON-decoded max width", t)
	assert.Equal(80, data.Profile.MaxHeight, "JSON-decoded max height", t)
	assert.Equal(int64(480), data.Profile.MaxArea, "JSON-decoded max area", t)

	// Make sure those profile variables are in the output data
	assert.True(bytes.Contains(w.Output, []byte("maxWidth")), "maxWidth", t)
	assert.True(bytes.Contains(w.Output, []byte("maxHeight")), "maxHeight", t)
	assert.True(bytes.Contains(w.Output, []byte("maxArea")), "maxArea", t)
}

// TestInfoNoMaxSize verifies that when the image is smaller than the handler's
// maximums, values are not present in the info profile
func TestInfoNoMaxSize(t *testing.T) {
	w := dorequest("docker%2Fimages%2Ftestfile%2Ftest-world-link.jp2/info.json", false, nc(6000, 8000, 4800000), t)
	var data iiif.Info
	json.Unmarshal(w.Output, &data)
	assert.Equal(0, data.Profile.MaxWidth, "JSON-decoded width", t)
	assert.Equal(0, data.Profile.MaxHeight, "JSON-decoded height", t)
	assert.Equal(int64(0), data.Profile.MaxArea, "JSON-decoded height", t)

	// Make sure those profile variables aren't in the output data at all
	assert.False(bytes.Contains(w.Output, []byte("maxWidth")), "no maxWidth", t)
	assert.False(bytes.Contains(w.Output, []byte("maxHeight")), "no maxHeight", t)
	assert.False(bytes.Contains(w.Output, []byte("maxArea")), "no maxArea", t)
}

func TestCommandHandler404(t *testing.T) {
	w := request("identifier/full/full/0/default.jpg", t)
	assert.Equal(404, w.StatusCode, "Valid command on nonexistent file returns 404", t)
}

func TestCommandHandlerInvalidFile(t *testing.T) {
	w := request("Makefile/full/full/0/default.jpg", t)
	assert.Equal(500, w.StatusCode, "Valid command on non-image file returns 500", t)
}

func TestInvalidRequest(t *testing.T) {
	w := request("docker%2Fimages%2Ftestfile%2Ftest-world.jp2/foo/bar", t)
	assert.Equal(400, w.StatusCode, "Bad request is reported as such", t)
}

func TestUnsupportedRequest(t *testing.T) {
	w := request("docker%2Fimages%2Ftestfile%2Ftest-world.jp2/pct:10,10,80,80/full/0/default.jpg", t)
	assert.Equal(501, w.StatusCode, "Unsupported operation gets reported as a 501 (not implemented)", t)
}

func TestCommandHandler(t *testing.T) {
	w := request("docker%2Fimages%2Ftestfile%2Ftest-world.jp2/10,10,80,80/full/0/default.jpg", t)
	assert.Equal(-1, w.StatusCode, "Valid command request doesn't explicitly set status code", t)
}

func TestCommandHandlerInvalidSize(t *testing.T) {
	imgid := "docker%2Fimages%2Ftestfile%2Ftest-world-link.jp2/pct:10,10,80,80/full/0/default.jpg"
	areaConstraint := nc(math.MaxInt32, math.MaxInt32, 480)
	wConstraint := nc(20, math.MaxInt32, math.MaxInt64)
	hConstraint := nc(math.MaxInt32, 20, math.MaxInt64)

	// For sanity let's make sure the request has no errors when we don't specify
	// any constraints
	w := dorequestl2(imgid, false, unlimited, t)
	assert.Equal(-1, w.StatusCode, "Supported request with valid size doesn't set status code", t)

	w = dorequestl2(imgid, false, wConstraint, t)
	assert.Equal(501, w.StatusCode, "Status code when width is too large", t)
	w = dorequestl2(imgid, false, hConstraint, t)
	assert.Equal(501, w.StatusCode, "Status code when height is too large", t)
	w = dorequestl2(imgid, false, areaConstraint, t)
	assert.Equal(501, w.StatusCode, "Status code when area is too large", t)
}

// BenchmarkRouting does a benchmark against the routing rules to ensure we
// aren't creating problems when changing how we interpret the incoming URLs.
func BenchmarkRouting(b *testing.B) {
	// Set up all the fake request bits outside the benchmark
	u, _ := url.Parse("http://example.com/foo/bar")
	w := fakehttp.NewResponseWriter()
	req, err := http.NewRequest("get", "", strings.NewReader(""))

	if err != nil {
		b.Errorf("Unable to create fake request: %s", err)
	}

	h := NewImageHandler(rootDir(), "/iiif")
	h.Maximums.Width = unlimited.Width
	h.Maximums.Height = unlimited.Height
	h.Maximums.Area = unlimited.Area
	h.BaseURL = u
	h.FeatureSet = iiif.FeatureSet2()
	URIs := []string{
		u.String() + "/foo/bar/invalid%2Fimage.jp2/10,10,80,80/full/0/default.jpg",
		u.String() + "/foo/bar/invalid%2Fimage.jp2/full/max/0/default.jpg",
		u.String() + "/foo/bar/invalid%2Fimage.jp2/full/pct:25/90/default.jpg",
		u.String() + "/foo/bar/invalid%2Fimage.jp2",
		u.String() + "/foo/bar/invalid%2Fimage.jp2/pct:10,10,900,900/max/0/default.png",
	}

	for n := 0; n < b.N; n++ {
		// We fire off fake requests of multiple types to test the different ways
		// the URL is parsed
		for _, requri := range URIs {
			req.RequestURI = requri
			h.IIIFRoute(w, req)
		}
	}
}

func TestIDToURL(t *testing.T) {
	var h = NewImageHandler("/var/local/images", "/iiif")
	h.AddSchemeMap("foo", "bar://real-host/prefixed-path")

	// Prefer table-driven tests, sirs
	var tests = map[string]struct {
		ID          string
		ExpectedURL *url.URL
	}{
		// iiif.ID should always be created by iiif.URLToID, so we don't need to
		// test out unescaped IDs here
		"simple": {
			"foo/bar/baz.jp2",
			&url.URL{Scheme: "file", Path: "/var/local/images/foo/bar/baz.jp2"},
		},
		"with scheme": {
			"s3://foo/bar/baz.jp2",
			&url.URL{Scheme: "s3", Host: "foo", Path: "/bar/baz.jp2"},
		},
		"explicit file won't resolve to an absolute path": {
			"file:///etc/passwd",
			&url.URL{Scheme: "file", Path: "/var/local/images/etc/passwd"},
		},
		"dot-dot problem": {
			"file:///../../../../../etc/passwd",
			&url.URL{Scheme: "file", Path: "/var/local/images/etc/passwd"},
		},
		"remapped scheme": {
			"foo://foo-host/foo-path/thing.jp2",
			&url.URL{Scheme: "bar", Host: "real-host", Path: "/prefixed-path/foo-host/foo-path/thing.jp2"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var got = h.getURL(iiif.ID(tc.ID))

			// We don't care about RawPath for testing purposes
			got.RawPath = ""

			var diff = cmp.Diff(tc.ExpectedURL, got)
			if diff != "" {
				t.Errorf("getURL(%q): %s", tc.ID, diff)
			}
		})
	}
}
