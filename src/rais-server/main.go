package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
)

var tilePath string

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var iiifURL string
	var address string

	flag.StringVar(&iiifURL, "iiif-url", "", `Base URL for serving IIIF requests, e.g., "http://example.com:8888/images/iiif"`)
	flag.StringVar(&address, "address", ":8888", "http service address")
	flag.StringVar(&tilePath, "tile-path", "", "Base path for JP2 images")
	flag.Parse()

	if tilePath == "" {
		fmt.Println("ERROR: --tile-path is required")
		flag.Usage()
		os.Exit(1)
	}

	http.HandleFunc("/images/tiles/", TileHandler)
	http.HandleFunc("/images/resize/", ResizeHandler)

	iiifBase, err := url.Parse(iiifURL)
	if iiifURL != "" && err != nil {
		fmt.Println("Invalid IIIF URL specified:", err)
		os.Exit(1)
	}

	if iiifBase.Scheme != "" && iiifBase.Host != "" && iiifBase.Path != "" {
		fmt.Printf("IIIF enabled at %s\n", iiifBase.String())
		ih := NewIIIFHandler(iiifBase, tilePath)
		http.HandleFunc(ih.Base.Path+"/", ih.Route)
	}

	if err := http.ListenAndServe(address, nil); err != nil {
		fmt.Printf("Error starting listener: %s", err)
		os.Exit(1)
	}
}
