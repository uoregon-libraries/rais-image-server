package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/hashicorp/golang-lru"
)

var tilePath string
var infoCache *lru.Cache

func main() {
	var iiifURL string
	var address string
	var infoCacheLen int

	flag.StringVar(&iiifURL, "iiif-url", "", `Base URL for serving IIIF requests, e.g., "http://example.com:8888/images/iiif"`)
	flag.StringVar(&address, "address", ":8888", "http service address")
	flag.StringVar(&tilePath, "tile-path", "", "Base path for JP2 images")
	flag.IntVar(&infoCacheLen, "iiif-info-cache-size", 10000, "Maximum cached image info entries (IIIF only)")
	flag.Parse()

	if tilePath == "" {
		log.Println("ERROR: --tile-path is required")
		flag.Usage()
		os.Exit(1)
	}

	http.HandleFunc("/images/tiles/", TileHandler)
	http.HandleFunc("/images/resize/", ResizeHandler)

	iiifBase, err := url.Parse(iiifURL)
	if iiifURL != "" && err != nil {
		log.Fatalf("Invalid IIIF URL specified: %s", err)
	}

	if iiifBase.Scheme != "" && iiifBase.Host != "" && iiifBase.Path != "" {
		if infoCacheLen > 0 {
			infoCache, err = lru.New(infoCacheLen)
			if err != nil {
				log.Fatalf("Unable to start info cache: %s", err)
			}
		}

		log.Printf("IIIF enabled at %s\n", iiifBase.String())
		ih := NewIIIFHandler(iiifBase, tilePath)
		http.HandleFunc(ih.Base.Path+"/", ih.Route)
	}

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Error starting listener: %s", err)
	}
}
