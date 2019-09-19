package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"net/http"
	"net/url"
	"rais/src/iiif"
	"rais/src/img"
	"rais/src/plugins"
	"strconv"
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

// ImageHandler responds to a IIIF URL request and parses the requested
// transformation within the limits of the handler's capabilities
type ImageHandler struct {
	IIIFBase   *url.URL
	FeatureSet *iiif.FeatureSet
	TilePath   string
	Maximums   img.Constraint
}

// NewImageHandler sets up a base ImageHandler with no features
func NewImageHandler(tp string) *ImageHandler {
	return &ImageHandler{
		TilePath: tp,
		Maximums: img.Constraint{Width: math.MaxInt32, Height: math.MaxInt32, Area: math.MaxInt64},
	}
}

// EnableIIIF sets up regexes for IIIF responses
func (ih *ImageHandler) EnableIIIF(u *url.URL) {
	ih.IIIFBase = u
	ih.FeatureSet = iiif.AllFeatures()
}

// cacheKey returns a key for caching if a given IIIF URL is cacheable by our
// current, somewhat restrictive, rules
func cacheKey(u *iiif.URL) string {
	if tileCache != nil && u.Format == iiif.FmtJPG && u.Size.W > 0 && u.Size.W <= 1024 && u.Size.H <= 1024 {
		return u.Path
	}
	return ""
}

// IIIFRoute takes an HTTP request and parses it to see what (if any) IIIF
// translation is requested
func (ih *ImageHandler) IIIFRoute(w http.ResponseWriter, req *http.Request) {
	// Pull identifier from base so we know if we're even dealing with a valid
	// file in the first place
	var url = *req.URL
	url.RawQuery = ""
	p := url.String()

	// If it didn't even match the base, something weird happened, so we just
	// spit out a generic 404
	prefix := ih.IIIFBase.Path + "/"
	if !strings.Contains(p, prefix) {
		http.NotFound(w, req)
		return
	}

	var urlPath = strings.Replace(p, prefix, "", 1)
	iiifURL, err := iiif.NewURL(urlPath)

	// If the iiifURL is invalid, it's possible this is a base URI request.  With
	// a bit of work we can find out if the path is a valid identifier.
	if err != nil {
		if ih.isValidBasePath(urlPath) {
			http.Redirect(w, req, p+"/info.json", 303)
		} else {
			http.Error(w, fmt.Sprintf("Invalid IIIF request %q: %s", iiifURL.Path, err), 400)
		}
		return
	}

	// Handle info.json prior to reading the image, in case of cached info
	fp := ih.getIIIFPath(iiifURL.ID)
	info, e := ih.getInfo(iiifURL.ID, fp)
	if e != nil {
		if e.Code != 404 {
			Logger.Errorf("Error getting IIIF info.json for resource %s (path %s): %s", iiifURL.ID, fp, e.Message)
		}
		http.Error(w, e.Message, e.Code)
		return
	}

	if iiifURL.Info {
		ih.Info(w, req, info)
		return
	}

	// Check the cache before spending the cycles to read in the image.  For now
	// the cache is very limited to ensure only relatively small requests are
	// actually cached.
	if key := cacheKey(iiifURL); key != "" {
		stats.TileCache.Get()
		data, ok := tileCache.Get(key)
		if ok {
			stats.TileCache.Hit()
			w.Header().Set("Content-Type", mime.TypeByExtension("."+string(iiifURL.Format)))
			w.Write(data.([]byte))
			return
		}
	}

	// No info path should mean a full command path - start reading the image
	res, err := img.NewResource(iiifURL.ID, fp)
	if err != nil {
		e := newImageResError(err)
		if e.Code != 404 {
			Logger.Errorf("Error initializing resource %s (path %s): %s", iiifURL.ID, fp, err)
		}
		http.Error(w, e.Message, e.Code)
		return
	}

	if !iiifURL.Valid() {
		// This means the URI was probably a command, but had an invalid syntax
		http.Error(w, "Invalid IIIF request: "+iiifURL.Error().Error(), 400)
		return
	}

	// Attempt to run the command
	ih.Command(w, req, iiifURL, res, info)
}

// isValidBasePath returns true if the given path is simply missing /info.json
// to function properly
func (ih *ImageHandler) isValidBasePath(path string) bool {
	var jsonPath = path + "/info.json"
	var iiifURL, err = iiif.NewURL(jsonPath)
	if err != nil {
		return false
	}

	var fp = ih.getIIIFPath(iiifURL.ID)
	var e *HandlerError
	_, e = ih.getInfo(iiifURL.ID, fp)
	return e == nil
}

func (ih *ImageHandler) getIIIFPath(id iiif.ID) string {
	for _, idtopath := range idToPathPlugins {
		fp, err := idtopath(id)
		if err == nil {
			return fp
		}
		if err == plugins.ErrSkipped {
			continue
		}
		Logger.Warnf("Error trying to use plugin to translate iiif.ID: %s", err)
	}
	return ih.TilePath + "/" + string(id)
}

func convertStrings(s1, s2, s3 string) (i1, i2, i3 int, err error) {
	i1, err = strconv.Atoi(s1)
	if err != nil {
		return
	}
	i2, err = strconv.Atoi(s2)
	if err != nil {
		return
	}
	i3, err = strconv.Atoi(s3)
	return
}

// Info responds to a IIIF info request with appropriate JSON based on the
// image's data and the handler's capabilities
func (ih *ImageHandler) Info(w http.ResponseWriter, req *http.Request, info *iiif.Info) {
	// Convert info to JSON
	json, err := marshalInfo(info)
	if err != nil {
		http.Error(w, err.Message, err.Code)
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

func newImageResError(err error) *HandlerError {
	switch err {
	case img.ErrDimensionsExceedLimits:
		return NewError(err.Error(), 501)
	case img.ErrDoesNotExist:
		return NewError("image resource does not exist", 404)
	default:
		return NewError(err.Error(), 500)
	}
}

func (ih *ImageHandler) getInfo(id iiif.ID, fp string) (info *iiif.Info, err *HandlerError) {
	// Check for cached image data first, and use that to create JSON
	info = ih.loadInfoFromCache(id)

	// Next, check for an overridden info.json file, and just spit that out
	// directly if it exists
	if info == nil {
		info = ih.loadInfoOverride(id, fp)
	}

	if info == nil {
		info, err = ih.loadInfoFromImageResource(id, fp)
	}

	return info, err
}

func (ih *ImageHandler) loadInfoFromCache(id iiif.ID) *iiif.Info {
	if infoCache == nil {
		return nil
	}

	stats.InfoCache.Get()
	data, ok := infoCache.Get(id)
	if !ok {
		return nil
	}

	stats.InfoCache.Hit()
	return ih.buildInfo(id, data.(ImageInfo))
}

func (ih *ImageHandler) loadInfoOverride(id iiif.ID, fp string) *iiif.Info {
	// If an override file isn't found or has an error, just skip it
	var infofile = fp + "-info.json"
	data, err := ioutil.ReadFile(infofile)
	if err != nil {
		return nil
	}

	Logger.Debugf("Loading image data from override file (%s)", infofile)

	// If an override file *is* found, replace the id
	escapedID := ih.IIIFBase.String() + "/" + id.Escaped()
	d2 := bytes.Replace(data, []byte("%ID%"), []byte(escapedID), 1)

	info := new(iiif.Info)
	err = json.Unmarshal(d2, info)
	if err != nil {
		Logger.Errorf("Cannot parse JSON override file %q: %s", infofile, err)
		return nil
	}
	return info
}

func (ih *ImageHandler) loadInfoFromImageResource(id iiif.ID, fp string) (*iiif.Info, *HandlerError) {
	Logger.Debugf("Loading image data from image resource (id: %s)", id)
	res, err := img.NewResource(id, fp)
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
		stats.InfoCache.Set()
		infoCache.Add(id, imageInfo)
	}
	return ih.buildInfo(id, imageInfo), nil
}

func (ih *ImageHandler) buildInfo(id iiif.ID, i ImageInfo) *iiif.Info {
	info := ih.FeatureSet.Info()
	info.Width = i.Width
	info.Height = i.Height

	if ih.Maximums.SmallerThanAny(i.Width, i.Height) {
		info.Profile.MaxArea = ih.Maximums.Area
		info.Profile.MaxWidth = ih.Maximums.Width
		info.Profile.MaxHeight = ih.Maximums.Height
	}

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
	info.ID = ih.IIIFBase.String() + "/" + id.Escaped()
	return info
}

func marshalInfo(info *iiif.Info) ([]byte, *HandlerError) {
	json, err := json.Marshal(info)
	if err != nil {
		Logger.Errorf("Unable to marshal IIIFInfo response: %s", err)
		return nil, NewError("server error", 500)
	}

	return json, nil
}

// Command handles image processing operations
func (ih *ImageHandler) Command(w http.ResponseWriter, req *http.Request, u *iiif.URL, res *img.Resource, info *iiif.Info) {
	// Send last modified time
	if err := sendHeaders(w, req, res.FilePath); err != nil {
		return
	}

	// Do we support this request?  If not, return a 501
	if !ih.FeatureSet.Supported(u) {
		http.Error(w, "Feature not supported", 501)
		return
	}

	var max = ih.Maximums

	// If we have an info, we can make use of it for the constraints rather than
	// using the global constraints; this is useful for overridden info.json files
	if info != nil {
		max = img.Constraint{
			Width:  info.Profile.MaxWidth,
			Height: info.Profile.MaxHeight,
			Area:   info.Profile.MaxArea,
		}
		if max.Width == 0 {
			max.Width = math.MaxInt32
		}
		if max.Height == 0 {
			max.Height = math.MaxInt32
		}
		if max.Area == 0 {
			max.Area = math.MaxInt64
		}
	}
	img, err := res.Apply(u, max)
	if err != nil {
		e := newImageResError(err)
		Logger.Errorf("Error applying transorm: %s", err)
		http.Error(w, e.Message, e.Code)
		return
	}

	w.Header().Set("Content-Type", mime.TypeByExtension("."+string(u.Format)))

	cacheBuf := bytes.NewBuffer(nil)
	if err := EncodeImage(cacheBuf, img, u.Format); err != nil {
		http.Error(w, "Unable to encode", 500)
		Logger.Errorf("Unable to encode to %s: %s", u.Format, err)
		return
	}

	if key := cacheKey(u); key != "" {
		stats.TileCache.Set()
		tileCache.Add(key, cacheBuf.Bytes())
	}

	if _, err := io.Copy(w, cacheBuf); err != nil {
		Logger.Errorf("Unable to encode to %s: %s", u.Format, err)
		return
	}
}
