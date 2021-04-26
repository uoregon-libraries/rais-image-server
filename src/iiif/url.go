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
	Path     string
	ID       ID
	Region   Region
	Size     Size
	Rotation Rotation
	Quality  Quality
	Format   Format
	Info     bool
}

type pathParts struct {
	data []string
}

func pathify(pth string) *pathParts {
	return &pathParts{data: strings.Split(pth, "/")}
}

// pop implements a hacky but fast and effective "pop" operation that just
// returns a blank string when there's nothing left to pop
func (p *pathParts) pop() string {
	var retval string
	if len(p.data) > 0 {
		retval, p.data = p.data[len(p.data)-1], p.data[:len(p.data)-1]
	}
	return retval
}

func (p *pathParts) rejoin() string {
	return strings.Join(p.data, "/")
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
//
// It's possible to get a URL and an error since an id-only request could
// theoretically exist for a resource with *any* id.  In those cases it's up to
// the caller to figure out what to do - the returned URL will have as much
// information as we're able to parse.
func NewURL(path string) (*URL, error) {
	var u = &URL{Path: path}

	// Check for an info request first since it's pretty trivial to do
	if strings.HasSuffix(path, "info.json") {
		u.Info = true
		u.ID = URLToID(strings.Replace(path, "/info.json", "", -1))
		return u, nil
	}

	// Parse in reverse order to deal with the continuing problem of slashes not
	// being escaped properly in all situations
	var parts = pathify(path)
	var qualityFormat = parts.pop()
	var qfParts = strings.SplitN(qualityFormat, ".", 2)
	if len(qfParts) == 2 {
		u.Format = StringToFormat(qfParts[1])
		u.Quality = StringToQuality(qfParts[0])
	}
	u.Rotation = StringToRotation(parts.pop())
	u.Size = StringToSize(parts.pop())
	u.Region = StringToRegion(parts.pop())

	// The remainder of the path has to be the ID
	u.ID = URLToID(parts.rejoin())

	// Invalid may or may not actually mean invalid, but we just let the caller
	// try to figure it out....
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
