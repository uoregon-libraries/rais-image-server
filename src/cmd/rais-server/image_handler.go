package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"net/http"
	"net/url"
	"path"
	"rais/src/iiif"
	"rais/src/img"
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
	BaseURL       *url.URL
	WebPathPrefix string
	FeatureSet    *iiif.FeatureSet
	TilePath      string
	Maximums      img.Constraint
}

// NewImageHandler sets up a base ImageHandler with no features
func NewImageHandler(tilePath, basePath string) *ImageHandler {
	return &ImageHandler{
		WebPathPrefix: basePath,
		TilePath:      tilePath,
		Maximums:      img.Constraint{Width: math.MaxInt32, Height: math.MaxInt32, Area: math.MaxInt64},
		FeatureSet:    iiif.AllFeatures(),
	}
}

// cacheKey returns a key for caching if a given IIIF URL is cacheable by our
// current, somewhat restrictive, rules
func cacheKey(u *iiif.URL) string {
	if tileCache != nil && u.Format == iiif.FmtJPG && u.Size.W > 0 && u.Size.W <= 1024 && u.Size.H <= 1024 {
		return u.Path
	}
	return ""
}

// getRequestURL determines the "real" request URL.  Proxies are supported by
// checking headers.  This should not be considered definitive - if RAIS is
// running standalone, users can fake these headers.  Fortunately, this is a
// read-only application, so it shouldn't be risky to rely on these headers as
// they only determine how to report the RAIS URLs in info.json requests.
func getRequestURL(req *http.Request) *url.URL {
	var host = req.Header.Get("X-Forwarded-Host")
	var scheme = req.Header.Get("X-Forwarded-Proto")
	if host != "" && scheme != "" {
		return &url.URL{Host: host, Scheme: scheme}
	}

	var u = &url.URL{
		Host:   req.Host,
		Scheme: "http",
	}
	if req.TLS != nil {
		u.Scheme = "https"
	}
	return u
}

// IIIFRoute takes an HTTP request and parses it to see what (if any) IIIF
// translation is requested
func (ih *ImageHandler) IIIFRoute(w http.ResponseWriter, req *http.Request) {
	// We need to take a copy of the URL, not the original, since we modify
	// things a bit
	var u = *req.URL

	// Massage the URL to ensure redirect / info requests will make sense
	u.RawQuery = ""
	u.Fragment = ""

	// Figure out the hostname, scheme, port, etc. either from the request or the
	// setting if it was explicitly set
	if ih.BaseURL != nil {
		u.Host = ih.BaseURL.Host
		u.Scheme = ih.BaseURL.Scheme
	} else {
		var u2 = getRequestURL(req)
		if u2 != nil {
			u.Host = u2.Host
			u.Scheme = u2.Scheme
		}
	}

	// Strip the IIIF web path off the beginning of the path to determine the
	// actual request.  This should always work because a request shouldn't be
	// able to get here if it didn't have our prefix.
	var prefix = ih.WebPathPrefix + "/"
	u.Path = strings.Replace(u.Path, prefix, "", 1)

	iiifURL, err := iiif.NewURL(u.Path)
	// If the iiifURL is invalid, it's possible this is a base URI request.
	// Let's see if treating the path as an ID gives us any info.
	if err != nil {
		if ih.isValidBasePath(u.Path) {
			http.Redirect(w, req, req.URL.String()+"/info.json", 303)
		} else {
			http.Error(w, fmt.Sprintf("Invalid IIIF request %q: %s", iiifURL.Path, err), 400)
		}
		return
	}

	// Grab the image resource and info data
	res, info, e := ih.getImageData(iiifURL.ID)
	if e != nil {
		if e.Code != 404 {
			Logger.Errorf("Error getting image and/or IIIF Info for %q: %s", iiifURL.ID, e.Message)
		}
		http.Error(w, e.Message, e.Code)
		return
	}

	defer res.Destroy()

	// Make sure the info JSON has the proper asset id, which, for some reason in
	// the IIIF spec, requires the full URL to the asset, not just its identifier
	infourl := &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   ih.WebPathPrefix,
	}

	// Because of how Go's URL path magic works, we really do have to just
	// concatenate these two things with a slash manually
	info.ID = infourl.String() + "/" + iiifURL.ID.Escaped()

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

	var res, _, e = ih.getImageData(iiifURL.ID)
	if res != nil {
		res.Destroy()
	}
	return e == nil
}

// getURL converts a IIIF ID into a URL.  If the ID has no scheme, we assume
// it's `file://`.  Additionally, all `file://` URIs get their path prefixed
// with the configured tilepath
func (ih *ImageHandler) getURL(id iiif.ID) *url.URL {
	// TODO: make this a plugin function, idToURL
	Logger.Warnf("add idToURL plugin hook")

	var u, err = url.Parse(string(id))
	// If an id fails to parse, it's probably a client-side error (such as
	// failing to escape the pound sign)
	if err != nil {
		u = &url.URL{Path: string(id)}
	}

	if u.Scheme == "" {
		u.Scheme = "file"
	}
	if u.Scheme == "file" {
		u.Path = path.Join(ih.TilePath, u.Path)
	}

	Logger.Debugf("%q translated to URL %q", id, u)

	return u
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
	if err == nil {
		return nil
	}

	// Allow wrapped errors for better messages without losing meanings
	if errors.Is(err, img.ErrDimensionsExceedLimits) {
		return NewError(err.Error(), 501)
	}
	if errors.Is(err, img.ErrDoesNotExist) {
		return NewError("image resource does not exist", 404)
	}

	// Unknown / unhandled errors are just general 500s
	return NewError(err.Error(), 500)
}

func (ih *ImageHandler) getImageData(id iiif.ID) (*img.Resource, *iiif.Info, *HandlerError) {
	var res, err = img.NewResource(id, ih.getURL(id))
	if err != nil {
		return nil, nil, newImageResError(err)
	}

	info, e := ih.getIIIFInfo(res)
	if e != nil {
		return nil, nil, e
	}

	return res, info, nil
}

func (ih *ImageHandler) getIIIFInfo(res *img.Resource) (*iiif.Info, *HandlerError) {
	// Check for cached image data first, and use that to create JSON
	var info = ih.loadInfoFromCache(res.ID)
	if info != nil {
		return info, nil
	}

	info = ih.loadInfoOverride(res)
	if info != nil {
		return info, nil
	}

	var err *HandlerError
	info, err = ih.loadInfoFromImageResource(res)
	if err != nil {
		return nil, err
	}

	return info, nil
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

func (ih *ImageHandler) loadInfoOverride(res *img.Resource) *iiif.Info {
	// If an override file isn't found or has an error, just skip it
	var infofile = res.URL.Path + "-info.json"
	var data, err = ioutil.ReadFile(infofile)
	if err != nil {
		return nil
	}

	Logger.Debugf("Loading image data from override file (%s)", infofile)

	var info = new(iiif.Info)
	err = json.Unmarshal(data, info)
	if err != nil {
		Logger.Errorf("Cannot parse JSON override file %q: %s", infofile, err)
		return nil
	}
	return info
}

func (ih *ImageHandler) saveInfoToCache(id iiif.ID, info ImageInfo) {
	if infoCache == nil {
		return
	}

	stats.InfoCache.Set()
	infoCache.Add(id, info)
}

func (ih *ImageHandler) loadInfoFromImageResource(res *img.Resource) (*iiif.Info, *HandlerError) {
	Logger.Debugf("Loading image data from image resource (id: %s)", res.ID)
	var d, err = res.Decoder()
	if err != nil {
		return nil, newImageResError(err)
	}

	var imageInfo = ImageInfo{
		Width:      d.GetWidth(),
		Height:     d.GetHeight(),
		TileWidth:  d.GetTileWidth(),
		TileHeight: d.GetTileHeight(),
		Levels:     d.GetLevels(),
	}

	// We save the minimal data to the cache so our cache remains incredibly
	// small for what it gives us
	ih.saveInfoToCache(res.ID, imageInfo)
	return ih.buildInfo(res.ID, imageInfo), nil
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
	if err := sendHeaders(w, req, res); err != nil {
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
