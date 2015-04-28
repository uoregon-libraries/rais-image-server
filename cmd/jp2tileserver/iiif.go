package main

import (
	"fmt"
	"regexp"
)

type Quality string
const (
	QColor   Quality = "color"
	QGray    Quality = "gray"
	QBitonal Quality = "bitonal"
	QDefault Quality = "default"
	QNative  Quality = "native"           // For 1.1 compatibility
)

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

type IIIFCommand struct {
	ID       string
	Region   Region
	Size     Size
	Rotation Rotation
	Quality  Quality
	Format   Format
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

func NewIIIFCommand(path string) *IIIFCommand {
	parts := iiifPathRegex.FindStringSubmatch(path)

	if parts == nil {
		return &IIIFCommand{}
	}

	iiif := &IIIFCommand{
		ID:       parts[1],
		Region:   StringToRegion(parts[2]),
		Size:     StringToSize(parts[3]),
		Rotation: StringToRotation(parts[4]),
		Quality:  Quality(parts[5]),
		Format:   Format(parts[6]),
	}

	return iiif
}

// Valid returns the validity of the request - if syntax is bad in any way
// (doesn't match the regex, region string violates syntax, etc), this returns
// false and the server should report a 400 status.
func (ic *IIIFCommand) Valid() bool {
	if ic.ID == "" || ic.Quality == Quality("") || ic.Format == Format("") {
		return false
	}

	if !ic.Region.Valid() || !ic.Size.Valid() {
		return false
	}

	return true
}
