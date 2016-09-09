package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iiif"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// AllFeatures is the complete list of everything supported by RAIS at this time
var AllFeatures = &iiif.FeatureSet{
	RegionByPx:  true,
	RegionByPct: true,

	SizeByWhListed: true,
	SizeByW:        true,
	SizeByH:        true,
	SizeByPct:      true,
	SizeByWh:       true,
	SizeByForcedWh: true,
	SizeAboveFull:  true,

	RotationBy90s:     true,
	RotationArbitrary: false,
	Mirroring:         true,

	Default: true,
	Color:   true,
	Gray:    true,
	Bitonal: true,

	Jpg:  true,
	Png:  true,
	Gif:  true,
	Tif:  true,
	Jp2:  false,
	Pdf:  false,
	Webp: false,

	BaseURIRedirect:     true,
	Cors:                true,
	JsonldMediaType:     true,
	ProfileLinkHeader:   false,
	CanonicalLinkHeader: false,
}

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

// IIIFHandler responds to an IIIF URL request and parses the requested
// transformation within the limits of the handler's capabilities
type IIIFHandler struct {
	Base          *url.URL
	BaseRegex     *regexp.Regexp
	BaseOnlyRegex *regexp.Regexp
	FeatureSet    *iiif.FeatureSet
	InfoPathRegex *regexp.Regexp
	TilePath      string
}

// NewIIIFHandler sets up an IIIFHandler with all features RAIS can support,
// listening based on the given base URL
func NewIIIFHandler(u *url.URL, tp string) *IIIFHandler {
	rprefix := fmt.Sprintf(`^%s`, u.Path)
	return &IIIFHandler{
		Base:          u,
		BaseRegex:     regexp.MustCompile(rprefix + `/([^/]+)`),
		BaseOnlyRegex: regexp.MustCompile(rprefix + `/[^/]+$`),
		InfoPathRegex: regexp.MustCompile(rprefix + `/([^/]+)/info.json$`),
		TilePath:      tp,
		FeatureSet:    AllFeatures,
	}
}

// Route takes an HTTP request and parses it to see what (if any) IIIF
// translation is requested
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

	id := iiif.ID(parts[1])
	fp := ih.TilePath + "/" + id.Path()

	// Check for base path and redirect if that's all we have
	if ih.BaseOnlyRegex.MatchString(p) {
		http.Redirect(w, req, p+"/info.json", 303)
		return
	}

	// Handle info.json prior to reading the image, in case of cached info
	if ih.InfoPathRegex.MatchString(p) {
		ih.Info(w, req, id, fp)
		return
	}

	// No info path should mean a full command path - start reading the image
	res, err := NewImageResource(id, fp)
	if err != nil {
		e := newImageResError(err)
		http.Error(w, e.Message, e.Code)
		return
	}

	u := iiif.NewURL(p)
	if !u.Valid() {
		// This means the URI was probably a command, but had an invalid syntax
		http.Error(w, "Invalid IIIF request", 400)
		return
	}

	// Attempt to run the command
	ih.Command(w, req, u, res)
}

// Info responds to a IIIF info request with appropriate JSON based on the
// image's data and the handler's capabilities
func (ih *IIIFHandler) Info(w http.ResponseWriter, req *http.Request, id iiif.ID, fp string) {
	// Check for cached image data first, and use that to create JSON
	json, e := ih.loadInfoJSONFromCache(id)
	if e != nil {
		http.Error(w, e.Message, e.Code)
		return
	}

	// Next, check for an overridden info.json file, and just spit that out
	// directly if it exists
	if json == nil {
		json = ih.loadInfoJSONOverride(id, fp)
	}

	if json == nil {
		json, e = ih.loadInfoJSONFromImageResource(id, fp)
		if e != nil {
			http.Error(w, e.Message, e.Code)
			return
		}
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

func newImageResError(err error) *HandlerError {
	switch err {
	case ErrImageDoesNotExist:
		return NewError("image resource does not exist", 404)
	default:
		return NewError(err.Error(), 500)
	}
}

func (ih *IIIFHandler) loadInfoJSONFromCache(id iiif.ID) ([]byte, *HandlerError) {
	if infoCache == nil {
		return nil, nil
	}

	data, ok := infoCache.Get(id)
	if !ok {
		return nil, nil
	}

	return ih.buildInfoJSON(id, data.(ImageInfo))
}

func (ih *IIIFHandler) loadInfoJSONOverride(id iiif.ID, fp string) []byte {
	// If an override file isn't found or has an error, just skip it
	json, err := ioutil.ReadFile(fp + "-info.json")
	if err != nil {
		return nil
	}

	// If an override file *is* found, replace the id
	fullid := ih.Base.String() + "/" + id.String()
	return bytes.Replace(json, []byte("%ID%"), []byte(fullid), 1)
}

func (ih *IIIFHandler) loadInfoJSONFromImageResource(id iiif.ID, fp string) ([]byte, *HandlerError) {
	log.Printf("Loading image data from image resource (id: %s)", id)
	res, err := NewImageResource(id, fp)
	if err != nil {
		return nil, newImageResError(err)
	}

	d := res.Decoder
	imageInfo := ImageInfo{
		Width:      d.GetWidth(),
		Height:     d.GetHeight(),
		TileWidth:  d.GetTileWidth(),
		TileHeight: d.GetTileHeight(),
		Levels:     d.GetLevels(),
	}

	if infoCache != nil {
		infoCache.Add(id, imageInfo)
	}
	return ih.buildInfoJSON(id, imageInfo)
}

func (ih *IIIFHandler) buildInfoJSON(id iiif.ID, i ImageInfo) ([]byte, *HandlerError) {
	info := ih.FeatureSet.Info()
	info.Width = i.Width
	info.Height = i.Height

	// Set up tile sizes
	if i.TileWidth > 0 {
		var sf []int
		scale := 1
		for x := 0; x < i.Levels; x++ {
			// For sanity's sake, let's not tell viewers they can get at absurdly
			// small sizes
			if info.Width/scale < 16 {
				break
			}
			if info.Height/scale < 16 {
				break
			}
			sf = append(sf, scale)
			scale <<= 1
		}
		info.Tiles = make([]iiif.TileSize, 1)
		info.Tiles[0] = iiif.TileSize{
			Width:        i.TileWidth,
			ScaleFactors: sf,
		}
		if i.TileHeight > 0 {
			info.Tiles[0].Height = i.TileHeight
		}
	}

	// The info id is actually the full URL to the resource, not just its ID
	info.ID = ih.Base.String() + "/" + id.String()

	json, err := json.Marshal(info)
	if err != nil {
		log.Printf("ERROR!  Unable to marshal IIIFInfo response: %s", err)
		return nil, NewError("server error", 500)
	}

	return json, nil
}

// Command handles image processing operations
func (ih *IIIFHandler) Command(w http.ResponseWriter, req *http.Request, u *iiif.URL, res *ImageResource) {
	// For now the cache is very limited to ensure only relatively small requests
	// are actually cached
	willCache := tileCache != nil && u.Format == iiif.FmtJPG && u.Size.W > 0 && u.Size.W <= 1024 && u.Size.H == 0
	cacheKey := u.Path

	// Send last modified time
	if err := sendHeaders(w, req, res.FilePath); err != nil {
		return
	}

	// Do we support this request?  If not, return a 501
	if !ih.FeatureSet.Supported(u) {
		http.Error(w, "Feature not supported", 501)
		return
	}

	// Check the cache now that we're sure everything is valid and supported.
	if willCache {
		data, ok := tileCache.Get(cacheKey)
		if ok {
			w.Header().Set("Content-Type", mime.TypeByExtension("."+string(u.Format)))
			w.Write(data.([]byte))
			return
		}
	}

	img, err := res.Apply(u)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", mime.TypeByExtension("."+string(u.Format)))

	var cacheBuf *bytes.Buffer
	var out io.Writer = w

	if willCache {
		cacheBuf = bytes.NewBuffer(nil)
		out = io.MultiWriter(w, cacheBuf)
	}

	if err := EncodeImage(out, img, u.Format); err != nil {
		http.Error(w, "Unable to encode", 500)
		log.Printf("Unable to encode to %s: %s", u.Format, err)
		return
	}

	if willCache {
		tileCache.Add(cacheKey, cacheBuf.Bytes())
	}
}
