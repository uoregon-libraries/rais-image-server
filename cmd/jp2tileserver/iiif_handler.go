package main

import (
	"encoding/json"
	"fmt"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/iiif"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/openjpeg"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/transform"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func acceptsLD(req *http.Request) bool {
	for _, h := range req.Header["Accept"] {
		for _, accept := range strings.Split(h, ",") {
			if accept == "application/ld+json" {
				return true
			}
		}
	}

	return false
}

type IIIFHandler struct {
	Base          *url.URL
	RegexPrefix   string
	BaseRegex     *regexp.Regexp
	BaseOnlyRegex *regexp.Regexp
	FeatureSet    *iiif.FeatureSet
	InfoPathRegex *regexp.Regexp
	TilePath      string
}

func NewIIIFHandler(u *url.URL, widths []int, tp string) *IIIFHandler {
	ih := &IIIFHandler{
		Base:        u,
		RegexPrefix: fmt.Sprintf(`^%s`, u.Path),
		TilePath:    tp,
		FeatureSet:  iiif.FeaturesLevel1,
	}

	ih.BaseRegex = regexp.MustCompile(ih.RegexPrefix + `/([^/]+)`)
	ih.BaseOnlyRegex = regexp.MustCompile(ih.RegexPrefix + `/[^/]+$`)
	ih.InfoPathRegex = regexp.MustCompile(ih.RegexPrefix + `/([^/]+)/info.json$`)

	// Add features we support beyond level 1
	fs := ih.FeatureSet
	fs.RotationBy90s = true

	// Tile sizes are set fairly statically
	fs.TileSizes = make([]iiif.TileSize, 0)
	sf := []int{1, 2, 4, 8, 16, 32, 64}
	for _, val := range widths {
		fs.TileSizes = append(fs.TileSizes, iiif.TileSize{Width: val, ScaleFactors: sf})
	}

	return ih
}

func (ih *IIIFHandler) Route(w http.ResponseWriter, req *http.Request) {
	// Pull identifier from base so we know if we're even dealing with a valid
	// file in the first place
	p := req.RequestURI
	parts := ih.BaseRegex.FindStringSubmatch(p)

	// If it didn't even match the base, something weird happened, so we just
	// spit out a generic 404
	if parts == nil {
		http.NotFound(w, req)
		return
	}

	identifier := iiif.ID(parts[1])
	filepath := ih.TilePath + "/" + identifier.Path()

	if _, err := os.Stat(filepath); err != nil {
		http.Error(w, "Image resource does not exist", 404)
		return
	}

	// Check for base path and redirect if that's all we have
	if ih.BaseOnlyRegex.MatchString(p) {
		http.Redirect(w, req, p+"/info.json", 303)
		return
	}

	// Check for info path, and dispatch if it matches
	if ih.InfoPathRegex.MatchString(p) {
		ih.Info(w, req, identifier)
		return
	}

	// No info path should mean a full command path
	if u := iiif.NewURL(p); u.Valid() {
		ih.Command(w, req, u)
		return
	}

	// This means the URI was probably a command, but had an invalid syntax
	http.Error(w, "Invalid IIIF request", 400)
}

func (ih *IIIFHandler) Info(w http.ResponseWriter, req *http.Request, id iiif.ID) {
	filepath := ih.TilePath + "/" + id.Path()
	jp2, err := openjpeg.NewJP2Image(filepath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read image %#v", id), 500)
		return
	}

	if err := jp2.ReadHeader(); err != nil {
		http.Error(w, fmt.Sprintf("Unable to read image dimensions for %#v", id), 500)
		return
	}

	info := ih.FeatureSet.Info()
	rect := jp2.Dimensions()
	info.Width = rect.Dx()
	info.Height = rect.Dy()

	// The info id is actually the full URL to the resource, not just its ID
	info.ID = ih.Base.String() + "/" + id.String()

	json, err := json.Marshal(info)
	if err != nil {
		log.Printf("ERROR!  Unable to marshal IIIFInfo response: %s", err)
		http.Error(w, "Server error", 500)
		return
	}

	// Set headers - content type is dependent on client
	ct := "application/json"
	if acceptsLD(req) {
		ct = "application/ld+json"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(json)
}

// Handles crop/resize operations for JP2s.  Note that this is the *wrong* way
// to handle most image formats.  JP2s are encoded as multi-resolution images,
// so the resize information actually has to be known before a given region can
// be cropped.  Otherwise we'd have to decode the whole image instead of just
// the minimum necessary "layer".
func (ih *IIIFHandler) Command(w http.ResponseWriter, req *http.Request, u *iiif.URL) {
	// Get file's last modified time, returning a 404 if we can't stat the file
	filepath := ih.TilePath + "/" + u.ID.Path()
	if err := sendHeaders(w, req, filepath); err != nil {
		return
	}

	// Do we support this request?  If not, return a 501
	if !ih.FeatureSet.Supported(u) {
		http.Error(w, "Feature not supported", 501)
		return
	}

	// Create JP2 structure - if we can't, the image must be corrupt or otherwise
	// broken, since we already checked for existence
	jp2, err := openjpeg.NewJP2Image(filepath)
	if err != nil {
		http.Error(w, "Unable to read source image", 500)
		log.Println("Unable to read source image: ", err)
		return
	}
	defer jp2.CleanupResources()

	if u.Region.Type == iiif.RTPixel {
		r := image.Rect(
			int(u.Region.X),
			int(u.Region.Y),
			int(u.Region.X+u.Region.W),
			int(u.Region.Y+u.Region.H),
		)
		jp2.SetCrop(r)
	}

	switch u.Size.Type {
	case iiif.STScaleToWidth:
		jp2.SetResizeWH(u.Size.W, 0)
	case iiif.STScaleToHeight:
		jp2.SetResizeWH(0, u.Size.H)
	case iiif.STExact:
		jp2.SetResizeWH(u.Size.W, u.Size.H)
	case iiif.STScalePercent:
		jp2.SetScale(u.Size.Percent / 100.0)
	}

	img, err := jp2.DecodeImage()
	if err != nil {
		http.Error(w, "Unable to decode image", 500)
		log.Println("Unable to decode image: ", err)
		return
	}

	if u.Rotation.Degrees != 0 {
		var r transform.Rotator
		switch img0 := img.(type) {
		case *image.Gray:
			r = transform.GrayRotator{img0}
		case *image.RGBA:
			r = transform.RGBARotator{img0}
		}

		switch u.Rotation.Degrees {
		case 90:
			img = r.Rotate90()
		case 180:
			img = r.Rotate180()
		case 270:
			img = r.Rotate270()
		}
	}

	// Encode as JPEG straight to the client
	if err = jpeg.Encode(w, img, &jpeg.Options{Quality: 80}); err != nil {
		http.Error(w, "Unable to encode jpeg", 500)
		log.Println("Unable to encode JPEG:", err)
		return
	}
}
