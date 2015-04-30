package main

import (
	"encoding/json"
	"fmt"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/iiif"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/openjpeg"
	"log"
	"net/http"
	"os"
	"regexp"
)

var iiifInfoPathRegex = regexp.MustCompile(`^/images/iiif/(.+)/info.json$`)

func IIIFHandler(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Path

	// Check for info path first, and dispatch if it matches
	if parts := iiifInfoPathRegex.FindStringSubmatch(p); parts != nil {
		iiifInfoHandler(w, req, parts[1])
		return
	}

	// No info path should mean a full command path
	if u := iiif.NewURL(p); u.Valid() {
		//iiifCommandHandler(w, req, u)
		return
	}

	// No info or command?  400 error according to IIIF spec.
	http.Error(w, "Invalid IIIF request", 400)
}

func iiifInfoHandler(w http.ResponseWriter, req *http.Request, id string) {
	fmt.Println(req.URL)

	filepath := tilePath + "/" + id
	if _, err := os.Stat(filepath); err != nil {
		http.Error(w, "Image resource does not exist", 404)
		return
	}

	jp2, err := openjpeg.NewJP2Image(filepath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read image %#v", id), 500)
		return
	}

	if err := jp2.ReadHeader(); err != nil {
		http.Error(w, fmt.Sprintf("Unable to read image dimensions for %#v", id), 500)
		return
	}

	info := NewIIIFInfo()
	rect := jp2.Dimensions()
	info.Width = rect.Dx()
	info.Height = rect.Dy()

	// The info id is actually the full path to the resource, not just its ID
	info.ID = iiifBase.String() + "/" + id

	json, err := json.Marshal(info)
	if err != nil {
		log.Printf("ERROR!  Unable to marshal IIIFInfo response: %s", err)
		http.Error(w, "Server error", 500)
		return
	}

	// Set headers - TODO: check for Accept header with jsonld content type!
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
