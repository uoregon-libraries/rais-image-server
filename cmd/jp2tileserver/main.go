package main

import (
	"flag"
	"fmt"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/openjpeg"
	"net/http"
	"os"
	"runtime"
)

var tilePath string

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
		fmt.Printf("Error starting listener: %s", err)
		os.Exit(1)
	}
}
