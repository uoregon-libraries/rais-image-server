// This file is an example of a plugin which would probably be a terrible idea,
// but should give devs a good feel for what can be done with a simple IDToPath
// implementation.
//
// When a resource is requested, if its IIIF id begins with "extern-http-" or
// "extern-https-", we treat "http" or "https" as the scheme and the rest of
// the id as the server, path, etc.  We download the image locally into /tmp if
// it has never been downloaded, convert it to a tiled, multi-resolution JP2,
// and return the path to said JP2.
//
// Note that for this example to work as-is, you must have the openjpeg
// compress tool in `/usr/bin/opj2_compress`.  This is not meant to be a
// real-world plugin with complex, configurable settings, so if you want to
// test it out, either run RAIS on a server where you can put the openjpeg
// binary there, or else run it using the RAIS dockerized "build box".
//
// This technique is extremely slow and error-prone, not to mention could
// easily burn an absurd amount of disk space.  But with some tweaks, a proper
// cache expiration setup, and some hostname limits, could allow for an image
// server that pulls internal images and presents them in a fast way (after the
// first hit, obviously).
//
// The obvious adaptations here could be things like s3 storage or even systems
// like Fedora Commons which hold files in a complicated filestore.  But an
// IDToPath plugin could also be as simple as taking opaque IDs and converting
// them into file paths for making a front-end application that isn't
// IIIF-aware easier to manage.

package main

import (
	"crypto/sha512"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"rais/src/iiif"
	"rais/src/plugins"
	"sync"

	"github.com/uoregon-libraries/gopkg/fileutil"
	"github.com/uoregon-libraries/gopkg/logger"
	"github.com/uoregon-libraries/gopkg/shell"
)

var m sync.Mutex

var l *logger.Logger

// SetLogger is called by the RAIS server's plugin manager to let plugins use
// the central logger
func SetLogger(raisLogger *logger.Logger) {
	l = raisLogger
}

// IDToPath implements the auto-download-and-convert logic when a IIIF ID
// starts with "extern-http" or "extern-https"
func IDToPath(id iiif.ID) (path string, err error) {
	var ids = string(id)
	if ids[:7] != "extern-" {
		return "", plugins.ErrSkipped
	}

	ids = ids[7:]
	if ids[:5] == "http-" {
		ids = "http://" + ids[5:]
	} else if ids[:6] == "https-" {
		ids = "https://" + ids[6:]
	} else {
		return "", plugins.ErrSkipped
	}

	// Check cache - don't re-download
	var hashed = hashName(id)
	path = hashed + ".jp2"

	// We don't want to pull multiple images at once, just as an extra safety
	// measure.  This should be a per-image measure, but this is good enough for
	// a simple external image example.
	m.Lock()
	defer m.Unlock()

	l.Debugf("external-images plugin: Checking for cached file at %q", path)
	if fileutil.MustNotExist(path) {
		err = pullImage(ids, hashed)
		if err == nil {
			err = convertImage(hashed, path)
		}
		os.Remove(hashed)
	} else {
		l.Debugf("external-images plugin: cached file found")
	}

	return path, err
}

// hashName returns a almost certainly unique string pointing to where a file
// will temporarily be stored based on its id
func hashName(id iiif.ID) string {
	var prefix = os.TempDir()
	var hash = sha512.Sum512([]byte(id))
	return filepath.Join(prefix, fmt.Sprintf("%x", hash))
}

// pullImage pulls the external file into a temporary cache, converts it to a
// JP2, and returns the path to that JP2
func pullImage(ids, path string) (err error) {
	var u string
	u, err = url.QueryUnescape(ids)
	if err != nil {
		return fmt.Errorf("external-images plugin: %s", err)
	}

	l.Infof("external-images plugin: Pulling file from %q", u)
	var resp *http.Response
	resp, err = http.Get(u)
	if err != nil {
		return fmt.Errorf("external-images plugin: %s", err)
	}
	defer resp.Body.Close()

	l.Debugf("external-images plugin: Writing file to %q", path)
	var f *os.File
	f, err = os.Create(path)
	if err != nil {
		return fmt.Errorf("external-images plugin: %s", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("external-images plugin: %s", err)
	}

	return nil
}

// convertImage runs a two-pass conversion on the given source image, first
// using imagemagick to convert it to a BMP, then converting the BMP to a JP2.
// This is an unfortunate necessity as `opj2_compress` doesn't handle very many
// image formats out of the box.  We could use imagemagick to convert directly
// to JP2, but we've found that to be somewhat less reliable in the past.
//
// We reconvert JP2s without checking, because not all JP2s are created
// equally: a JP2 that isn't tiled or multi-resolution isn't going to be as
// performant as one which is both.  Additionally, a lossless JP2 can use a lot
// more space than one that's high quality, but lossy.  We could try to check
// all these factors, but it's not really worthwhile for a prototype.
func convertImage(src, dest string) error {
	var bmpPath = src + ".bmp"
	var ok = shell.ExecSubgroup("/usr/bin/convert", src, bmpPath)
	if !ok {
		return fmt.Errorf("external-images plugin: unable to convert to BMP")
	}

	defer os.Remove(bmpPath)

	var jp2Path = src + ".jp2"
	ok = shell.ExecSubgroup("/usr/bin/opj2_compress", "-i", bmpPath, "-o", jp2Path, "-t", "1024,1024", "-r", "20.250")
	if !ok {
		return fmt.Errorf("external-images plugin: unable to convert BMP to JP2")
	}

	return nil
}
