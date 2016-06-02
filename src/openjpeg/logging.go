package openjpeg

// #cgo pkg-config: libopenjp2
// #include "handlers.h"
import "C"

import (
	"fmt"
	"strings"
)

// LogLevel is the hard-coded log level.  It's forcibly set to WARN.  This
// whole thing needs a rewrite.
var LogLevel = 4

// LogLevels contains all possible levels for GoLog and goLog
var LogLevels = []string{"EMERG", "ALERT", "CRIT", "ERROR", "WARN", "NOTICE", "INFO", "DEBUG"}

// GoLog bridges the openjpeg logging with our internal logger
//export GoLog
func GoLog(clevel C.int, cmessage *C.char) {
	level := int(clevel)
	message := C.GoString(cmessage)

	goLog(level, "FROM OPJ: "+strings.TrimSpace(message))
}

// Internal go-specific version of logger
func goLog(level int, message string) {
	if level <= LogLevel {
		fmt.Printf("[%s] %s\n", LogLevels[level], message)
	}
}
