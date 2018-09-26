package iiif

import (
	"errors"
	"net/url"
	"strings"
)

// ID is a string identifying a particular file to process.  It should be
// unescaped for use if it's coming from a URL via URLToID().
type ID string

// URLToID converts a value pulled from a URL into a suitable IIIF ID (by unescaping it)
func URLToID(val string) ID {
	s, _ := url.QueryUnescape(string(val))
	return ID(s)
}

// Escaped returns an escaped version of the ID, suitable for use in a URL
func (id ID) Escaped() string {
	return url.QueryEscape(string(id))
}

// URL represents the different options composed into a IIIF URL request
type URL struct {
	Path            string
	ID              ID
	Region          Region
	Size            Size
	Rotation        Rotation
	Quality         Quality
	Format          Format
	Info            bool
	BaseURIRedirect bool
}

// NewURL takes a path string (no scheme, server, or prefix, just the IIIF
// pieces), such as "path%2Fto%2Fsomefile.jp2/full/512,/270/default.jpg", and
// breaks it down into the different components.  In this example:
//
//     - ID:       "path%2Fto%2Fsomefile.jp2" (the server determines how to find the image)
//     - Region:   "full"                     (the whole image is processed)
//     - Size:     "512,"                     (the image is resized to a width of 512; aspect ratio is maintained)
//     - Rotation: "270"                      (the image is rotated 270 degrees clockwise)
//     - Quality:  "default"                  (the image color space is unchanged)
//     - Format:   "jpg"                      (the resulting image will be a JPEG)
func NewURL(path string) (*URL, error) {
	u := &URL{Path: path}

	parts := strings.Split(path, "/")

	// First check for an ID-only request, which must be redirected
	if len(parts) == 1 {
		u.ID = URLToID(path)
		u.Info = true
		u.BaseURIRedirect = true
		u.Path = ""
		return u, nil
	}

	// Now check for an info request
	last := len(parts) - 1
	qualityFormat := parts[last]
	if len(parts) == 2 && qualityFormat == "info.json" {
		u.Info = true
		u.ID = URLToID(parts[0])
		return u, nil
	}

	// All requests for images must have exactly 5 parts
	if len(parts) != 5 {
		return u, errors.New("invalid IIIF path components")
	}

	qfParts := strings.SplitN(qualityFormat, ".", 2)
	if len(qfParts) != 2 {
		return u, errors.New("invalid quality/format specifier")
	}

	u.ID = URLToID(parts[0])
	u.Region = StringToRegion(parts[1])
	u.Size = StringToSize(parts[2])
	u.Rotation = StringToRotation(parts[3])
	u.Quality = StringToQuality(qfParts[0])
	u.Format = StringToFormat(qfParts[1])

	if !u.Valid() {
		return u, u.Error()
	}

	return u, nil
}

// Valid returns the validity of the request - is the syntax is bad in any way?
// Are any numbers outside a set range?  Was the identifier blank?  Etc.
//
// Invalid requests are expected to report an http status of 400.
func (u *URL) Valid() bool {
	return u.Error() == nil
}

// Error returns an error specifying invalid parts of the URL
func (u *URL) Error() error {
	var messages []string
	if u.ID == "" {
		messages = append(messages, "empty id")
	}
	if !u.Region.Valid() {
		messages = append(messages, "invalid region")
	}
	if !u.Size.Valid() {
		messages = append(messages, "invalid size")
	}
	if !u.Rotation.Valid() {
		messages = append(messages, "invalid rotation")
	}
	if !u.Quality.Valid() {
		messages = append(messages, "invalid quality")
	}
	if !u.Format.Valid() {
		messages = append(messages, "invalid format")
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, ", "))
	}
	return nil
}
