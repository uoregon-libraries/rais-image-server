package openjpeg

// #cgo pkg-config: libopenjp2
// #include "handlers.h"
import "C"

import (
	"strings"

	"github.com/uoregon-libraries/gopkg/logger"
)

// Logger defaults to use a default implementation of the uoregon-libraries
// logging mechanism, but can be overridden (as is the case with the main RAIS
// command)
var Logger = logger.Named("rais/openjpeg", logger.Debug)

// GoLogWarning bridges the openjpeg logging with our internal logger
//export GoLogWarning
func GoLogWarning(cmessage *C.char) {
	log(Logger.Warnf, cmessage)
}

// GoLogError bridges the openjpeg logging with our internal logger
//export GoLogError
func GoLogError(cmessage *C.char) {
	log(Logger.Errorf, cmessage)
}

// Internal go-specific version of logger
func log(logfn func(string, ...interface{}), cmessage *C.char) {
	var message = strings.TrimSpace(C.GoString(cmessage))
	logfn("FROM OPJ: %s", message)
}
