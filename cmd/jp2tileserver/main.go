package main

import (
	"flag"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"fmt"
	"os"
	"time"

	"github.com/nfnt/resize"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/openjpeg"
)

var tilePathRegex = regexp.MustCompile(`/images/tiles/(?P<path>.+)/image_(?P<width>\d+)x(?P<height>\d+)_from_(?P<x1>\d+),(?P<y1>\d+)_to_(?P<x2>\d+),(?P<y2>\d+).jpg`)
var resizePathRegex = regexp.MustCompile(`/images/resize/(.+)/(\d+)x(\d+)`)
var infoPathRegex = regexp.MustCompile(`^/images/info/(.+)$`)

var tilePath string

func InfoHandler(w http.ResponseWriter, req *http.Request) {
	// Extract request path's regex parts into local variables
	parts := infoPathRegex.FindStringSubmatch(req.URL.Path)

	if parts == nil {
		http.Error(w, "Invalid info request", 400)
		return
	}

	path := parts[1]
	filepath := tilePath + "/" + path
	jp2 := openjpeg.NewJP2Image(filepath)
	rect, err := jp2.Dimensions()
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read JP2 dimensions for %#v", path), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"size": [%d, %d]}`, rect.Dx(), rect.Dy())
}

func sendCacheHeaders(w http.ResponseWriter, filepath string) error {
	info, err := os.Stat(filepath)
	if err != nil {
		http.Error(w, "Unable to access file", 404)
		return err
	}

	// Set headers
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Last-Modified", info.ModTime().Format(time.RFC1123))
	return nil
}

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

	if err := sendCacheHeaders(w, filepath); err != nil {
		return
	}

	// Create JP2 structure
	jp2 := openjpeg.NewJP2Image(filepath)
	defer jp2.CleanupResources()

	// Pull raw tile data
	jp2.SetResize(width, height)
	jp2.SetCrop(r)
	img, err := jp2.RawImage()
	if err != nil {
		http.Error(w, "Unable to read source image", 500)
		log.Println("Unable to read source image: ", err)
		return
	}

	// Encode as JPEG straight to the client
	if err = jpeg.Encode(w, img, &jpeg.Options{Quality: 80}); err != nil {
		http.Error(w, "Unable to encode tile", 500)
		log.Println("Unable to encode tile into JPEG:", err)
		return
	}
}

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

	if err := sendCacheHeaders(w, filepath); err != nil {
		http.Error(w, "Unable to access file " + path, 404)
		return
	}

	// Create JP2 structure
	jp2 := openjpeg.NewJP2Image(filepath)
	defer jp2.CleanupResources()

	// Pull raw tile data
	jp2.SetResize(width, height)
	img, err := jp2.RawImage()
	if err != nil {
		http.Error(w, "Unable to read source image", 500)
		log.Println("Unable to read source image: ", err)
		return
	}

	// Encode as JPEG straight to the client
	resized := resize.Resize(uint(width), uint(height), img, resize.Bilinear)
	if err = jpeg.Encode(w, resized, &jpeg.Options{Quality: 80}); err != nil {
		http.Error(w, "Unable to encode tile", 500)
		log.Println("Unable to encode tile into JPEG:", err)
		return
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var address string
	var logLevel int

	flag.StringVar(&address, "address", ":8888", "http service address")
	flag.StringVar(&tilePath, "tile-path", "", "Base path for JP2 images")
	flag.IntVar(&logLevel, "log-level", 4, "Log level: 0-7 (lower is less verbose)")
	flag.Parse()

	if tilePath == "" {
		fmt.Println("ERROR: --tile-path is required")
		flag.Usage()
		os.Exit(1)
	}

	openjpeg.LogLevel = logLevel

	http.HandleFunc("/images/tiles/", TileHandler)
	http.HandleFunc("/images/info/", InfoHandler)
	http.HandleFunc("/images/resize/", ResizeHandler)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Print("ListenAndServe:", err)
	}
}
