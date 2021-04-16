// Package magick is a hacked up port of the minimal functionality we need
// to satisfy the img.Decoder interface.  Code is based in part on
// github.com/quirkey/magick
package main

/*
#cgo pkg-config: MagickCore
#include <magick/MagickCore.h>
*/
import "C"
import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"rais/src/img"
	"rais/src/plugins"
	"strings"
	"unsafe"

	"github.com/uoregon-libraries/gopkg/logger"
)

var l *logger.Logger

// SetLogger is called by the RAIS server's plugin manager to let plugins use
// the central logger
func SetLogger(raisLogger *logger.Logger) {
	l = raisLogger
}

// Initialize sets up the MagickCore stuff and registers the TIFF, PNG, JPG,
// and GIF decoders
func Initialize() {
	path, _ := os.Getwd()
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	C.MagickCoreGenesis(cPath, C.MagickFalse)
	C.SetMagickResourceLimit(C.DiskResource, C.MagickResourceInfinity)
	img.RegisterDecodeHandler(decodeCommonFile)
}

func makeError(where string, exception *C.ExceptionInfo) error {
	var reason = C.GoString(exception.reason)
	var description = C.GoString(exception.description)
	return fmt.Errorf("ImageMagick/%s: API Error #%v: %q - %q", where, exception.severity, reason, description)
}

var validExtensions = []string{".tif", ".tiff", ".png", ".jpg", ".jpeg", ".gif"}

func validExt(u *url.URL) bool {
	var ext = strings.ToLower(filepath.Ext(u.Path))
	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}

	return false
}

func validScheme(u *url.URL) bool {
	return u.Scheme == "file"
}

func decodeCommonFile(s img.Streamer) (img.DecodeFunc, error) {
	var u = s.Location()
	if !validExt(u) {
		l.Debugf("plugins/imagick-decoder: skipping unsupported image extension %q (must be one of %s)",
			s.Location(), strings.Join(validExtensions, ", "))
		return nil, plugins.ErrSkipped
	}

	if !validScheme(u) {
		l.Debugf("plugins/imagick-decoder: skipping unsupported URL scheme %q (must be file)", u.Scheme)
		return nil, plugins.ErrSkipped
	}

	return func() (img.Decoder, error) { return NewImage(u.Path) }, nil
}
