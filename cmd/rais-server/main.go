package main

import (
	"flag"
	"fmt"
	"github.com/uoregon-libraries/rais-image-server/openjpeg"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var tilePath string

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var tileSizeString, iiifURL string
	var address string
	var logLevel int

	flag.StringVar(&iiifURL, "iiif-url", "", `Base URL for serving IIIF requests, e.g., "http://example.com:8888/images/iiif"`)
	flag.StringVar(&tileSizeString, "iiif-tile-sizes", "512", `Tile sizes for IIIF, e.g., "256,512,1024"`)
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
	http.HandleFunc("/images/resize/", ResizeHandler)

	iiifBase, err := url.Parse(iiifURL)
	if iiifURL != "" && err != nil {
		fmt.Println("Invalid IIIF URL specified:", err)
		os.Exit(1)
	}

	if iiifBase.Scheme != "" && iiifBase.Host != "" && iiifBase.Path != "" {
		fmt.Printf("IIIF enabled at %s\n", iiifBase.String())
		tileSizes := parseInts(tileSizeString)
		ih := NewIIIFHandler(iiifBase, tileSizes, tilePath)
		http.HandleFunc(ih.Base.Path+"/", ih.Route)
	}

	if err := http.ListenAndServe(address, nil); err != nil {
		fmt.Printf("Error starting listener: %s", err)
		os.Exit(1)
	}
}

func parseInts(intStrings string) []int {
	iList := make([]int, 0)

	if intStrings == "" {
		return iList
	}

	for _, s := range strings.Split(intStrings, ",") {
		i, err := strconv.Atoi(s)
		if err != nil {
			fmt.Println("-- Error parsing tile sizes; using default of 512")
			return []int{512}
		}

		if i > 0 {
			iList = append(iList, i)
		}
	}

	return iList
}
