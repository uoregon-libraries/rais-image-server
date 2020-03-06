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
	"os"
	"path/filepath"
	"rais/src/img"
	"rais/src/plugins"
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

func decodeCommonFile(path string) (img.DecodeFunc, error) {
	switch filepath.Ext(path) {
	case ".tif", ".tiff", ".png", ".jpg", "jpeg", ".gif":
		return func() (img.Decoder, error) { return NewImage(path) }, nil
	default:
		return nil, plugins.ErrSkipped
	}
}
