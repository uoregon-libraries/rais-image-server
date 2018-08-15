package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"rais/src/fakehttp"
	"rais/src/iiif"
	"strings"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
	"github.com/uoregon-libraries/gopkg/logger"
)

func init() {
	Logger = logger.New(logger.Warn)
}

func rootDir() string {
	p, _ := os.Getwd()
	root, _ := filepath.Abs(p + "/../../../")
	return root
}

// Sets up everything necessary to test an IIIF request
func dorequest(path string, acceptLD bool, t *testing.T) *fakehttp.ResponseWriter {
	u, _ := url.Parse("http://example.com/foo/bar")
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
	h := NewImageHandler(rootDir())
	h.EnableIIIF(u)
	h.FeatureSet = iiif.FeatureSet1()
	h.IIIFRoute(w, req)

	return w
}

func request(path string, t *testing.T) *fakehttp.ResponseWriter {
	return dorequest(path, false, t)
}

func requestLD(path string, t *testing.T) *fakehttp.ResponseWriter {
	return dorequest(path, true, t)
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
	assert.Equal("http://iiif.io/api/image/2/level2.json", data.Profile[0], "Proper profile string", t)
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
	assert.Equal("http://iiif.io/api/image/2/level1.json", data.Profile[0], "Proper profile string", t)
	assert.Equal(800, data.Width, "JSON-decoded width", t)
	assert.Equal(400, data.Height, "JSON-decoded height", t)
	assert.Equal(800, data.Tiles[0].Width, "JSON-decoded tile width", t)
	assert.Equal(400, data.Tiles[0].Height, "JSON-decoded tile height", t)
	assert.Equal(1, len(data.Tiles[0].ScaleFactors), "1 scale factor exists", t)
	assert.Equal(1, data.Tiles[0].ScaleFactors[0], "Scale factor is 1", t)
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
