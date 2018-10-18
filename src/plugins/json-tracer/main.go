// This file creates a plugin for instrumenting RAIS for internal use.  It
// provides similar information as the DataDog plugin, but in a more "raw" way.
// Usage will require a new configuration value, "TracerOutputDirectory", or an
// environment value in RAIS_TRACEROUTPUTDIRECTORY.  To avoid docker-compose
// file proliferation, this plugin doesn't provide an example for stringing
// together docker-compose.blah.yml files.  To use this with docker-compose,
// do something like this in docker-compose.override.yml:
//
// environment:
//   - RAIS_TRACEROUTPUTDIRECTORY=/tmp/rais-traces
//   - RAIS_TRACEFLUSHMINUTES=10
//   - RAIS_MAXTRACES=10000

package main

import (
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/uoregon-libraries/gopkg/logger"
)

var l *logger.Logger
var jsonOutDir string
var reg *registry

// Disabled lets the plugin manager know not to add this plugin's functions to
// the global list unless sanity checks in Initialize() pass
var Disabled = true

// flushTime is the duration after which traces are flushed to disk
var flushTime time.Duration

// maxTraces is used to forcibly flush data at a certain point even if the
// elapsed flushTime hasn't passed
var maxTraces int

// Initialize reads configuration and sets up the JSON output directory
func Initialize() {
	viper.SetDefault("TraceFlushMinutes", 10)
	viper.SetDefault("MaxTraces", 10000)

	jsonOutDir = viper.GetString("TracerOutputDirectory")
	var flushMinutes = viper.GetInt("TraceFlushMinutes")
	maxTraces = viper.GetInt("MaxTraces")
	flushTime = time.Minute * time.Duration(flushMinutes)

	if jsonOutDir == "" {
		l.Warnf("TracerOutputDirectory must be configured, or RAIS_TRACEROUTPUTDIRECTORY must be set in the environment  **JSON Tracer plugin is disabled**")
		return
	}

	var err = os.MkdirAll(jsonOutDir, 0750)
	if err != nil {
		l.Errorf("json-tracer plugin: unable to create json tracer output directory %q: %s", jsonOutDir, err)
		return
	}

	reg = new(registry)

	Disabled = false
}

// WrapHandler takes all RAIS routes' handlers and wraps them with the JSON
// tracer middleware
func WrapHandler(pattern string, handler http.Handler) (http.Handler, error) {
	return reg.new(handler), nil
}

// SetLogger is called by the RAIS server's plugin manager to let plugins use
// the central logger
func SetLogger(raisLogger *logger.Logger) {
	l = raisLogger
}

// Teardown writes all pending information to the JSON directory
func Teardown() {
	reg.shutdown()
}
