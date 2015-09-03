package iiif

import (
	"fmt"
	"net/url"
	"regexp"
)

// ID is a string identifying a particular file to process.  It can contain
// URI-encoded data in order to allow, e.g., full paths.
type ID string

// Path unescapes "percentage encoding" to return a more friendly value for
// path-on-disk usage.
func (id ID) Path() string {
	p, _ := url.QueryUnescape(string(id))
	return p
}

// String just gives the ID as it was created, but obviously as a string type
func (id ID) String() string {
	return string(id)
}

var pathRegex = regexp.MustCompile(fmt.Sprintf(
	"/%s/%s/%s/%s/%s.%s$",
	`([^/]+)`,
	`(full|\d+,\d+,\d+,\d+|pct:[0-9.]+,[0-9.]+,[0-9.]+,[0-9.]+)`,
	`(full|\d+,|,\d+|pct:[0-9.]+|\d+,\d+|!\d+,\d+)`,
	`(\d+|!\d+)`,
	`(color|gray|bitonal|default|native)`,
	`(jpg|tif|png|gif|jp2|pdf|webp)`,
))

// URL represents the different options composed into an IIIF URL request
type URL struct {
	ID       ID
	Region   Region
	Size     Size
	Rotation Rotation
	Quality  Quality
	Format   Format
}

// NewURL takes a path string (no scheme, server, or prefix, just the IIIF
// pieces), such as "somefile.jp2/full/512,/270/default.jpg", and breaks it
// down into the different components.  In this example:
//
//     - ID:       "somefile.jp2"  (the server determines how to find this)
//     - Region:   "full"          (the whole image is processed)
//     - Size:     "512,"          (the image is resized to a width of 512; aspect ratio is maintained)
//     - Rotation: "270"           (the image is rotated 270 degrees clockwise)
//     - Quality:  "default"       (the image color space is unchanged)
//     - Format:   "jpg"           (the resulting image will be a JPEG)
func NewURL(path string) *URL {
	parts := pathRegex.FindStringSubmatch(path)

	if parts == nil {
		return &URL{}
	}

	return &URL{
		ID:       ID(parts[1]),
		Region:   StringToRegion(parts[2]),
		Size:     StringToSize(parts[3]),
		Rotation: StringToRotation(parts[4]),
		Quality:  Quality(parts[5]),
		Format:   Format(parts[6]),
	}
}

// Valid returns the validity of the request - is the syntax is bad in any way?
// Are any numbers outside a set range?  Was the identifier blank?  Etc.
//
// Invalid requests are expected to report an http status of 400.
func (u *URL) Valid() bool {
	return u.ID != "" &&
		u.Region.Valid() &&
		u.Size.Valid() &&
		u.Rotation.Valid() &&
		u.Quality.Valid() &&
		u.Format.Valid()
}
