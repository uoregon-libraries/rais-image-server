package main

import (
	"image"
	"image/jpeg"
	"net/http"
	"regexp"
	"strconv"
)

var tilePathRegex = regexp.MustCompile(`^/images/tiles/(?P<path>.+)/image_(?P<width>\d+)x(?P<height>\d+)_from_(?P<x1>\d+),(?P<y1>\d+)_to_(?P<x2>\d+),(?P<y2>\d+).jpg`)
var resizePathRegex = regexp.MustCompile(`^/images/resize/(.+)/(\d+)x(\d+)`)
var infoPathRegex = regexp.MustCompile(`^/images/info/(.+)$`)

// TileHandler is responsible for all chronam-like tile requests
func TileHandler(w http.ResponseWriter, req *http.Request) {
	// Extract request path's regex parts into local variables
	parts := tilePathRegex.FindStringSubmatch(req.URL.Path)

	if parts == nil {
		http.Error(w, "Invalid tile request", 400)
		return
	}

	d := map[string]string{}
	for i, name := range tilePathRegex.SubexpNames() {
		d[name] = parts[i]
	}
	path := d["path"]
	x1, _ := strconv.Atoi(d["x1"])
	y1, _ := strconv.Atoi(d["y1"])
	x2, _ := strconv.Atoi(d["x2"])
	y2, _ := strconv.Atoi(d["y2"])
	r := image.Rect(x1, y1, x2, y2)
	width, _ := strconv.Atoi(d["width"])
	height, _ := strconv.Atoi(d["height"])

	filepath := tilePath + "/" + path

	if err := sendHeaders(w, req, filepath); err != nil {
		return
	}

	res, err := NewImageResource("", filepath)
	if err != nil {
		http.Error(w, "Unable to read source image", 500)
		logger.Errorf("Unable to read source image: %s", err)
		return
	}
	i := res.Decoder

	// Pull raw tile data
	i.SetResizeWH(width, height)
	i.SetCrop(r)
	img, err := i.DecodeImage()
	if err != nil {
		http.Error(w, "Unable to decode image", 500)
		logger.Errorf("Unable to decode image: %s", err)
		return
	}

	// Encode as JPEG straight to the client
	w.Header().Set("Content-Type", "image/jpeg")
	if err = jpeg.Encode(w, img, &jpeg.Options{Quality: 80}); err != nil {
		http.Error(w, "Unable to encode tile", 500)
		logger.Errorf("Unable to encode tile into JPEG: %s", err)
		return
	}
}

// ResizeHandler is responsible for all chronam-like resize requests
func ResizeHandler(w http.ResponseWriter, req *http.Request) {
	// Extract request path's regex parts into local variables
	parts := resizePathRegex.FindStringSubmatch(req.URL.Path)
	if parts == nil {
		http.Error(w, "Invalid resize request", 400)
		return
	}

	path := parts[1]
	width, _ := strconv.Atoi(parts[2])
	height, _ := strconv.Atoi(parts[3])

	// Get file's last modified time, returning a 404 if we can't stat the file
	filepath := tilePath + "/" + path

	if err := sendHeaders(w, req, filepath); err != nil {
		return
	}

	res, err := NewImageResource("", filepath)
	if err != nil {
		http.Error(w, "Unable to read source image", 500)
		logger.Errorf("Unable to read source image: %s", err)
		return
	}
	i := res.Decoder

	// Pull raw tile data
	i.SetResizeWH(width, height)
	img, err := i.DecodeImage()
	if err != nil {
		http.Error(w, "Unable to decode image", 500)
		logger.Errorf("Unable to decode image: %s", err)
		return
	}

	// Encode as JPEG straight to the client
	w.Header().Set("Content-Type", "image/jpeg")
	if err = jpeg.Encode(w, img, &jpeg.Options{Quality: 80}); err != nil {
		http.Error(w, "Unable to encode tile", 500)
		logger.Errorf("Unable to encode tile into JPEG: %s", err)
		return
	}
}
