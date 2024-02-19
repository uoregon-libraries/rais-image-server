// This file creates a plugin for instrumenting RAIS for internal use.  It
// provides similar information as the DataDog plugin, but in a more "raw" way.
// Usage will require a new configuration value, "TracerOut", or an environment
// value in RAIS_TRACEROUT.
//
// To avoid docker-compose file proliferation, this plugin doesn't provide an
// example for stringing together docker-compose.blah.yml files.  To use this
// with docker-compose, do something like this in docker-compose.override.yml:
//
// environment:
//   - RAIS_TRACEROUT=/tmp/rais-traces.json
//   - RAIS_TRACERFLUSHSECONDS=10

package main

import (
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/uoregon-libraries/gopkg/logger"
)

var l *logger.Logger
var jsonOut string
var reg *registry

// Disabled lets the plugin manager know not to add this plugin's functions to
// the global list unless sanity checks in Initialize() pass
var Disabled = true

// flushTime is the duration after which events are flushed to disk
var flushTime time.Duration

// Initialize reads configuration and sets up the JSON output directory
func Initialize() {
	viper.SetDefault("TracerFlushSeconds", 10)
	flushTime = time.Second * time.Duration(viper.GetInt("TracerFlushSeconds"))
	jsonOut = viper.GetString("TracerOut")

	if jsonOut == "" {
		l.Warnf("TracerOut must be configured, or RAIS_TRACEROUT must be set in the environment  **JSON Tracer plugin is disabled**")
		return
	}

	reg = new(registry)

	Disabled = false
}

// WrapHandler takes all RAIS routes' handlers and wraps them with the JSON
// tracer middleware
func WrapHandler(_ string, handler http.Handler) (http.Handler, error) {
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
