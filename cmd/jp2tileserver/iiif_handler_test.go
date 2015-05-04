package main

import (
	"encoding/json"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/color-assert"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/fakehttp"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func rootDir() string {
	p, _ := os.Getwd()
	root, _ := filepath.Abs(p + "/../../")
	return root
}

// Sets up everything necessary to test an IIIF request
func request(path string, t *testing.T) *fakehttp.ResponseWriter {
	tilePath = rootDir()
	iiifBase, _ = url.Parse("http://example.com/images/iiif")
	w := fakehttp.NewResponseWriter()
	req, err := http.NewRequest("get", path, strings.NewReader(""))
	if err != nil {
		t.Errorf("Unable to create fake request: %s", err)
	}
	IIIFHandler(w, req)

	return w
}

func TestInfoHandler404(t *testing.T) {
	w := request("/images/iiif/identifier/info.json", t)
	assert.Equal(404, w.StatusCode, "Invalid info request returns 404", t)
}

func TestInfoHandler(t *testing.T) {
	w := request("/images/iiif/test-world.jp2/info.json", t)
	assert.Equal(-1, w.StatusCode, "Valid info request doesn't explicitly set status code", t)
	var data IIIFInfo
	json.Unmarshal(w.Output, &data)
	assert.Equal(800, data.Width, "JSON-decoded width", t)
	assert.Equal(400, data.Height, "JSON-decoded height", t)
	assert.Equal("http://example.com/images/iiif/test-world.jp2", data.ID, "JSON-decoded ID", t)
	assert.Equal(1, len(w.Headers["Content-Type"]), "Proper content type length", t)
	assert.Equal("application/json", w.Headers["Content-Type"][0], "Proper content type", t)
}

func TestCommandHandler404(t *testing.T) {
	w := request("/images/iiif/identifier/full/full/0/default.jpg", t)
	assert.Equal(404, w.StatusCode, "Valid command on nonexistent file returns 404", t)
}

func TestCommandHandlerInvalidFile(t *testing.T) {
	w := request("/images/iiif/Makefile/full/full/0/default.jpg", t)
	assert.Equal(500, w.StatusCode, "Valid command on non-image file returns 500", t)
}

func TestInvalidRequest(t *testing.T) {
	w := request("/images/iiif/test-world.jp2/foo/bar", t)
	assert.Equal(400, w.StatusCode, "Bad request is reported as such", t)
}

func TestUnsupportedRequest(t *testing.T) {
	w := request("/images/iiif/test-world.jp2/pct:10,10,80,80/full/0/default.jpg", t)
	assert.Equal(501, w.StatusCode, "Unsupported operation gets reported as a 501 (not implemented)", t)
}

func TestCommandHandler(t *testing.T) {
	w := request("/images/iiif/test-world.jp2/10,10,80,80/full/0/default.jpg", t)
	assert.Equal(-1, w.StatusCode, "Valid command request doesn't explicitly set status code", t)
}
