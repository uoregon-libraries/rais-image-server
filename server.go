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

	"github.com/eikeon/brikker/openjpeg"
)

var e = regexp.MustCompile(`/images/tiles/(?P<path>.+)/image_(?P<width>\d+)x(?P<height>\d+)_from_(?P<x1>\d+),(?P<y1>\d+)_to_(?P<x2>\d+),(?P<y2>\d+).jpg`)

func TileHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/jpeg")
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
	if err, i := openjpeg.NewImageTile(path, r, width, height); err == nil {
		if err = jpeg.Encode(w, i, nil); err != nil {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	Address := flag.String("address", ":8888", "http service address")
	flag.Parse()

	http.Handle("/", http.HandlerFunc(TileHandler))

	if err := http.ListenAndServe(*Address, nil); err != nil {
		log.Print("ListenAndServe:", err)
	}
}
