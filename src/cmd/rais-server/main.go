package main

import (
	"net/http"
	"net/url"
	"rais/src/iiif"
	"rais/src/magick"
	"rais/src/openjpeg"
	"rais/src/plugins"
	"rais/src/version"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/golang-lru"
	"github.com/spf13/viper"
	"github.com/uoregon-libraries/gopkg/interrupts"
	"github.com/uoregon-libraries/gopkg/logger"
)

var tilePath string
var infoCache *lru.Cache
var tileCache *lru.TwoQueueCache

// Logger is the server's central logger.Logger instance
var Logger *logger.Logger

const defaultAddress = ":12415"
const defaultInfoCacheLen = 10000

// cacheHits and cacheMisses allow some rudimentary tracking of cache value
var cacheHits, cacheMisses int64

var defaultLogLevel = logger.Debug.String()
var defaultPlugins = "s3-images.so,json-tracer.so"

func main() {
	parseConf()

	Logger = logger.New(logger.LogLevelFromString(viper.GetString("LogLevel")))
	openjpeg.Logger = Logger
	magick.Logger = Logger

	var plugPatterns = strings.Split(viper.GetString("Plugins"), ",")
	LoadPlugins(Logger, plugPatterns)

	tilePath = viper.GetString("TilePath")
	address := viper.GetString("Address")

	ih := NewImageHandler(tilePath)
	ih.Maximums.Area = viper.GetInt64("ImageMaxArea")
	ih.Maximums.Width = viper.GetInt("ImageMaxWidth")
	ih.Maximums.Height = viper.GetInt("ImageMaxHeight")

	iiifBase, _ := url.Parse(viper.GetString("IIIFURL"))

	icl := viper.GetInt("InfoCacheLen")
	if icl > 0 {
		infoCache, err = lru.New(icl)
		if err != nil {
			Logger.Fatalf("Unable to start info cache: %s", err)
		}
	}

	tcl := viper.GetInt("TileCacheLen")
	if tcl > 0 {
		Logger.Debugf("Creating a tile cache to hold up to %d tiles", tcl)
		tileCache, err = lru.New2Q(tcl)
		if err != nil {
			Logger.Fatalf("Unable to start info cache: %s", err)
		}
	}

	Logger.Infof("IIIF enabled at %s", iiifBase.String())
	ih.EnableIIIF(iiifBase)

	capfile := viper.GetString("CapabilitiesFile")
	if capfile != "" {
		ih.FeatureSet = &iiif.FeatureSet{}
		_, err := toml.DecodeFile(capfile, &ih.FeatureSet)
		if err != nil {
			Logger.Fatalf("Invalid file or formatting in capabilities file '%s'", capfile)
		}
		Logger.Debugf("Setting IIIF capabilities from file '%s'", capfile)
	}

	handle(ih.IIIFBase.Path+"/", http.HandlerFunc(ih.IIIFRoute))
	handle("/images/dzi/", http.HandlerFunc(ih.DZIRoute))
	handle("/version", http.HandlerFunc(VersionHandler))

	Logger.Infof("RAIS v%s starting...", version.Version)
	var srv = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		Addr:         address,
	}

	var wait sync.WaitGroup

	interrupts.TrapIntTerm(func() {
		wait.Add(1)
		Logger.Infof("Stopping RAIS...")
		srv.Shutdown(nil)

		if len(teardownPlugins) > 0 {
			Logger.Infof("Tearing down plugins")
			for _, plug := range teardownPlugins {
				plug()
			}
			Logger.Infof("Plugin teardown complete")
		}

		Logger.Infof("Stopped")
		wait.Done()
	})

	if err := srv.ListenAndServe(); err != nil {
		// Don't report a fatal error when we close the server
		if err != http.ErrServerClosed {
			Logger.Fatalf("Error starting listener: %s", err)
		}
	}
	wait.Wait()
}

// handle sends the pattern and raw handler to plugins, and sets up routing on
// whatever is returned (if anything).  All plugins which wrap handlers are
// allowed to run, but the behavior could definitely get weird depending on
// what a given plugin does.  Ye be warned.
func handle(pattern string, handler http.Handler) {
	for _, plug := range wrapHandlerPlugins {
		var h2, err = plug(pattern, handler)
		if err == nil {
			handler = h2
		} else if err != plugins.ErrSkipped {
			logger.Fatalf("Error trying to wrap handler %q: %s", pattern, err)
		}
	}
	http.Handle(pattern, handler)
}

// VersionHandler spits out the raw version string to the browser
func VersionHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(version.Version))
}
