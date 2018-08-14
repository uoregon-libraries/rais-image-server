package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"rais/src/iiif"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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

// DZITileSize defines deep zoom tile size
const DZITileSize = 1024

// DZI "constants" - these can be declared once (unlike IIIF regexes) because
// we aren't allowing a custom DZI handler path
var (
	DZIInfoRegex     = regexp.MustCompile(`^/images/dzi/(.+).dzi$`)
	DZITilePathRegex = regexp.MustCompile(`^/images/dzi/(.+)_files/(\d+)/(\d+)_(\d+).jpg$`)
)

// ImageHandler responds to an IIIF URL request and parses the requested
// transformation within the limits of the handler's capabilities
type ImageHandler struct {
	IIIFBase          *url.URL
	IIIFBaseRegex     *regexp.Regexp
	IIIFBaseOnlyRegex *regexp.Regexp
	IIIFInfoPathRegex *regexp.Regexp
	FeatureSet        *iiif.FeatureSet
	TilePath          string
}

// NewImageHandler sets up a base ImageHandler with no features
func NewImageHandler(tp string) *ImageHandler {
	return &ImageHandler{TilePath: tp}
}

// EnableIIIF sets up regexes for IIIF responses
func (ih *ImageHandler) EnableIIIF(u *url.URL) {
	rprefix := fmt.Sprintf(`^%s`, u.Path)
	ih.IIIFBase = u
	ih.IIIFBaseRegex = regexp.MustCompile(rprefix + `/([^/]+)`)
	ih.IIIFBaseOnlyRegex = regexp.MustCompile(rprefix + `/[^/]+$`)
	ih.IIIFInfoPathRegex = regexp.MustCompile(rprefix + `/([^/]+)/info.json$`)
	ih.FeatureSet = AllFeatures
}

// IIIFRoute takes an HTTP request and parses it to see what (if any) IIIF
// translation is requested
func (ih *ImageHandler) IIIFRoute(w http.ResponseWriter, req *http.Request) {
	// Pull identifier from base so we know if we're even dealing with a valid
	// file in the first place
	var url = *req.URL
	url.RawQuery = ""
	p := url.String()
	parts := ih.IIIFBaseRegex.FindStringSubmatch(p)

	// If it didn't even match the base, something weird happened, so we just
	// spit out a generic 404
	if parts == nil {
		http.NotFound(w, req)
		return
	}

	id := iiif.ID(parts[1])
	fp := ih.TilePath + "/" + id.Path()

	// Check for base path and redirect if that's all we have
	if ih.IIIFBaseOnlyRegex.MatchString(p) {
		http.Redirect(w, req, p+"/info.json", 303)
		return
	}

	// Handle info.json prior to reading the image, in case of cached info
	if ih.IIIFInfoPathRegex.MatchString(p) {
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

// DZIRoute takes an HTTP request and parses for responding to DZI info and
// tile requests
func (ih *ImageHandler) DZIRoute(w http.ResponseWriter, req *http.Request) {
	p := req.RequestURI

	var id iiif.ID
	var filePath string
	var handler func(*ImageResource)

	parts := DZIInfoRegex.FindStringSubmatch(p)
	if parts != nil {
		id = iiif.ID(parts[1])
		filePath = ih.TilePath + "/" + id.Path()
		handler = func(r *ImageResource) { ih.DZIInfo(w, r) }
	}

	parts = DZITilePathRegex.FindStringSubmatch(p)
	if parts != nil {
		id = iiif.ID(parts[1])
		filePath = ih.TilePath + "/" + id.Path()

		level, tileX, tileY, err := convertStrings(parts[2], parts[3], parts[4])
		if err != nil {
			http.Error(w, "Invalid DZI request", 400)
			return
		}

		handler = func(r *ImageResource) { ih.DZITile(w, req, r, level, tileX, tileY) }
	}

	if handler == nil {
		// Some kind of invalid request - just spit out a generic 404
		http.NotFound(w, req)
		return
	}

	res, err := NewImageResource(id, filePath)
	if err != nil {
		e := newImageResError(err)
		http.Error(w, e.Message, e.Code)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	handler(res)
}

// DZIInfo returns the info response for a deep-zoom request.  This is very
// hard-coded because the XML is simple, and deep-zoom is really a minor
// addition to RAIS.
func (ih *ImageHandler) DZIInfo(w http.ResponseWriter, res *ImageResource) {
	format := `<?xml version="1.0" encoding="UTF-8"?>
		<Image xmlns="http://schemas.microsoft.com/deepzoom/2008" TileSize="%d" Overlap="0" Format="jpg">
			<Size Width="%d" Height="%d"/>
		</Image>`
	d := res.Decoder
	xml := fmt.Sprintf(format, DZITileSize, d.GetWidth(), d.GetHeight())
	w.Write([]byte(xml))
}

// DZITile returns a tile by setting up the appropriate IIIF data structure
// based on the deep-zoom request
func (ih *ImageHandler) DZITile(w http.ResponseWriter, req *http.Request, res *ImageResource, l, tX, tY int) {
	// We always return 256x256 or bigger
	if l < 8 {
		l = 8
	}

	// Figure out max level
	d := res.Decoder
	srcWidth := uint64(d.GetWidth())
	srcHeight := uint64(d.GetHeight())

	var maxDimension uint64
	if srcWidth > srcHeight {
		maxDimension = srcWidth
	} else {
		maxDimension = srcHeight
	}
	maxLevel := uint64(math.Ceil(math.Log2(float64(maxDimension))))

	// Ignore absurd levels - even above 20 is pretty unlikely, but this is
	// called "future-proofing".  Or something.
	if uint64(l) > maxLevel {
		http.Error(w, fmt.Sprintf("Image doesn't support zoom level %d", l), 400)
		return
	}

	// Construct an IIIF URL so we can just reuse the IIIF-centric Command function
	var reduction uint64 = 1 << (maxLevel - uint64(l))

	width := reduction * DZITileSize
	height := reduction * DZITileSize
	left := uint64(tX) * width
	top := uint64(tY) * height

	// Fail early if the tile requested isn't legit
	if tX < 0 || tY < 0 || left > srcWidth || top > srcHeight {
		http.Error(w, "Invalid tile request", 400)
		return
	}

	// If any dimension is outside the image area, we have to adjust the requested image and tilesize
	finalWidth := width
	finalHeight := height
	if left+width > srcWidth {
		finalWidth = srcWidth - left
	}
	if top+height > srcHeight {
		finalHeight = srcHeight - top
	}

	requestedWidth := DZITileSize * finalWidth / width

	u := iiif.NewURL(fmt.Sprintf("%s/%d,%d,%d,%d/%d,/0/default.jpg",
		res.FilePath, left, top, finalWidth, finalHeight, requestedWidth))
	ih.Command(w, req, u, res)
}

// Info responds to a IIIF info request with appropriate JSON based on the
// image's data and the handler's capabilities
func (ih *ImageHandler) Info(w http.ResponseWriter, req *http.Request, id iiif.ID, fp string) {
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

func (ih *ImageHandler) loadInfoJSONFromCache(id iiif.ID) ([]byte, *HandlerError) {
	if infoCache == nil {
		return nil, nil
	}

	data, ok := infoCache.Get(id)
	if !ok {
		return nil, nil
	}

	return ih.buildInfoJSON(id, data.(ImageInfo))
}

func (ih *ImageHandler) loadInfoJSONOverride(id iiif.ID, fp string) []byte {
	// If an override file isn't found or has an error, just skip it
	json, err := ioutil.ReadFile(fp + "-info.json")
	if err != nil {
		return nil
	}

	// If an override file *is* found, replace the id
	fullid := ih.IIIFBase.String() + "/" + id.String()
	return bytes.Replace(json, []byte("%ID%"), []byte(fullid), 1)
}

func (ih *ImageHandler) loadInfoJSONFromImageResource(id iiif.ID, fp string) ([]byte, *HandlerError) {
	Logger.Debugf("Loading image data from image resource (id: %s)", id)
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

func (ih *ImageHandler) buildInfoJSON(id iiif.ID, i ImageInfo) ([]byte, *HandlerError) {
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
	info.ID = ih.IIIFBase.String() + "/" + id.String()

	json, err := json.Marshal(info)
	if err != nil {
		Logger.Errorf("Unable to marshal IIIFInfo response: %s", err)
		return nil, NewError("server error", 500)
	}

	return json, nil
}

// Command handles image processing operations
func (ih *ImageHandler) Command(w http.ResponseWriter, req *http.Request, u *iiif.URL, res *ImageResource) {
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
		Logger.Errorf("Unable to encode to %s: %s", u.Format, err)
		return
	}

	if willCache {
		tileCache.Add(cacheKey, cacheBuf.Bytes())
	}
}
