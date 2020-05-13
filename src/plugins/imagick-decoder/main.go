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
	img.RegisterDecodeHandler(decodeCommonFile)
}

func makeError(exception *C.ExceptionInfo) error {
	return fmt.Errorf("%v: %v - %v", exception.severity, exception.reason, exception.description)
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
		l.Infof("plugins/imagick-decoder: skipping unsupported image extension %q (must be one of %s)",
			s.Location(), strings.Join(validExtensions, ", "))
		return nil, plugins.ErrSkipped
	}

	// This is sorta of overly "loud" (warning), but generally speaking, a
	// decoder shouldn't be requiring local files, so we want people to be made
	// aware this plugin's not great....
	if !validScheme(u) {
		l.Warnf("plugins/imagick-decoder: skipping unsupported URL scheme %q (must be file)")
		return nil, plugins.ErrSkipped
	}

	return func() (img.Decoder, error) { return NewImage(u.Path) }, nil
}
