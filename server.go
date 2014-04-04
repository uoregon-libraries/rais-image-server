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

	"github.com/eikeon/brikker/openjpeg"
)

var e = regexp.MustCompile(`/images/tiles/(?P<path>.+)/image_(?P<width>\d+)x(?P<height>\d+)_from_(?P<x1>\d+),(?P<y1>\d+)_to_(?P<x2>\d+),(?P<y2>\d+).jpg`)

var tilePath string

func TileHandler(w http.ResponseWriter, req *http.Request) {
	// Extract request path's regex parts into local variables
	parts := e.FindStringSubmatch(req.URL.Path)
	d := map[string]string{}
	for i, name := range e.SubexpNames() {
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

	// Get file's last modified time, returning a 404 if we can't stat the file
	filepath := tilePath + "/" + path
	info, err := os.Stat(filepath)
	if err != nil {
		http.NotFound(w, req)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Last-Modified", info.ModTime().Format(time.RFC1123))

	// Serve generated JPG file
	if err, i := openjpeg.NewImageTile(filepath, r, width, height); err == nil {
		if err = jpeg.Encode(w, i, nil); err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var address string

	flag.StringVar(&address, "address", ":8888", "http service address")
	flag.StringVar(&tilePath, "tile-path", "", "Base path for JP2 images")
	flag.Parse()

	if tilePath == "" {
		fmt.Println("ERROR: --tile-path is required")
		flag.Usage()
		os.Exit(1)
	}

	http.Handle("/", http.HandlerFunc(TileHandler))

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Print("ListenAndServe:", err)
	}
}
