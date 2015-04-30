package main

import (
	"flag"
	"fmt"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/openjpeg"
	"net/http"
	"net/url"
	"os"
	"runtime"
)

var tilePath string
var iiifBase *url.URL

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var iiifScheme, iiifServer, iiifPrefix string
	var address string
	var logLevel int

	flag.StringVar(&iiifScheme, "iiif-scheme", "", `Scheme for serving IIIF requests, e.g., "http"`)
	flag.StringVar(&iiifServer, "iiif-server", "", `Server for serving IIIF requests, e.g., "example.com:8888"`)
	flag.StringVar(&iiifPrefix, "iiif-prefix", "", `Prefix for serving IIIF requests, e.g., "images/iiif"`)
	flag.StringVar(&address, "address", ":8888", "http service address")
	flag.StringVar(&tilePath, "tile-path", "", "Base path for JP2 images")
	flag.IntVar(&logLevel, "log-level", 4, "Log level: 0-7 (lower is less verbose)")
	flag.Parse()

	iiifBase = &url.URL{
		Scheme: iiifScheme,
		Host:   iiifServer,
		Path:   iiifPrefix,
	}

	if tilePath == "" {
		fmt.Println("ERROR: --tile-path is required")
		flag.Usage()
		os.Exit(1)
	}

	openjpeg.LogLevel = logLevel

	http.HandleFunc("/images/tiles/", TileHandler)
	http.HandleFunc("/images/info/", InfoHandler)
	http.HandleFunc("/images/resize/", ResizeHandler)

	if iiifBase.Scheme != "" && iiifBase.Host != "" && iiifBase.Path != "" {
		fmt.Printf("IIIF enabled at %s\n", iiifBase.String())
		http.HandleFunc("/"+iiifBase.Path+"/", IIIFHandler)
	}

	if err := http.ListenAndServe(address, nil); err != nil {
		fmt.Printf("Error starting listener: %s", err)
		os.Exit(1)
	}
}
