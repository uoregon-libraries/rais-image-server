package iiif

import (
	"fmt"
	"net/url"
	"regexp"
)

type ID string

func (id ID) Path() string {
	p, _ := url.QueryUnescape(string(id))
	return p
}

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

type URL struct {
	ID       ID
	Region   Region
	Size     Size
	Rotation Rotation
	Quality  Quality
	Format   Format
}

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
