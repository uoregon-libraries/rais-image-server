package iiif

import (
	"fmt"
	"net/url"
	"regexp"
)

type Quality string

const (
	QColor   Quality = "color"
	QGray    Quality = "gray"
	QBitonal Quality = "bitonal"
	QDefault Quality = "default"
	QNative  Quality = "native" // For 1.1 compatibility
)

var Qualities = []Quality{QColor, QGray, QBitonal, QDefault, QNative}

func (q Quality) Valid() bool {
	for _, valid := range Qualities {
		if valid == q {
			return true
		}
	}

	return false
}

type Format string

const (
	FmtJPG  Format = "jpg"
	FmtTIF  Format = "tif"
	FmtPNG  Format = "png"
	FmtGIF  Format = "gif"
	FmtJP2  Format = "jp2"
	FmtPDF  Format = "pdf"
	FmtWEBP Format = "webp"
)

var Formats = []Format{FmtJPG, FmtTIF, FmtPNG, FmtJP2, FmtPDF, FmtWEBP}

func (f Format) Valid() bool {
	for _, valid := range Formats {
		if valid == f {
			return true
		}
	}

	return false
}

type URL struct {
	ID       ID
	Region   Region
	Size     Size
	Rotation Rotation
	Quality  Quality
	Format   Format
}

type ID string

func (id ID) Path() string {
	p, _ := url.QueryUnescape(string(id))
	return p
}

func (id ID) String() string {
	return string(id)
}


var iiifPathRegex = regexp.MustCompile(fmt.Sprintf(
	"/%s/%s/%s/%s/%s.%s$",
	`([^/]+)`,                                                    // identifier
	`(full|\d+,\d+,\d+,\d+|pct:[0-9.]+,[0-9.]+,[0-9.]+,[0-9.]+)`, // region
	`(full|\d+,|,\d+|pct:[0-9.]+|\d+,\d+|!\d+,\d+)`,              // size
	`(\d+|!\d+)`,                                                 // rotation
	`(color|gray|bitonal|default|native)`,                        // quality
	`(jpg|tif|png|gif|jp2|pdf|webp)`,                             // format
))

func NewURL(path string) *URL {
	parts := iiifPathRegex.FindStringSubmatch(path)

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
