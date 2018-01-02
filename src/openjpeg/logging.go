package openjpeg

// #cgo pkg-config: libopenjp2
// #include "handlers.h"
import "C"

import (
	"strings"

	"github.com/uoregon-libraries/gopkg/logger"
)

// Logger defaults to use the standard uoregon-libraries logging mechanism, but
// can be overridden (as is the case with the main RAIS command)
var Logger = logger.DefaultLogger

// GoLogInfo bridges the openjpeg logging with our internal logger.  We map
// openjpeg's "info" logging to debug because these tend to be very verbose and
// not terribly important.
//export GoLogInfo
func GoLogInfo(cmessage *C.char) {
	log(Logger.Debugf, cmessage)
}

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
