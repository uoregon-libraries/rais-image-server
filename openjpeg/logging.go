package openjpeg

// #cgo LDFLAGS: -lopenjp2
// #include "handlers.h"
import "C"

import (
	"fmt"
	"strings"
)

// WARN by default
var LogLevel = 4
var LogLevels = []string{"EMERG", "ALERT", "CRIT", "ERROR", "WARN", "NOTICE", "INFO", "DEBUG"}

//export GoLog
func GoLog(clevel C.int, cmessage *C.char) {
	level := int(clevel)
	message := C.GoString(cmessage)

	goLog(level, "FROM OPJ: " + strings.TrimSpace(message))
}

// Internal go-specific version of logger
func goLog(level int, message string) {
	if level <= LogLevel {
		fmt.Printf("[%s] %s\n", LogLevels[level], message)
	}
}
