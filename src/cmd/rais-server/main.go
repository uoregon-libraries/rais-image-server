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

var infoCache *lru.Cache
var tileCache *lru.TwoQueueCache

// Logger is the server's central logger.Logger instance
var Logger *logger.Logger

// Global server stats for admin information gathering
var stats = new(serverStats)

func main() {
	parseConf()
	Logger = logger.New(logger.LogLevelFromString(viper.GetString("LogLevel")))
	openjpeg.Logger = Logger
	magick.Logger = Logger

	setupCaches()
	LoadPlugins(Logger, strings.Split(viper.GetString("Plugins"), ","))

	tilePath := viper.GetString("TilePath")
	address := viper.GetString("Address")

	ih := NewImageHandler(tilePath)
	ih.Maximums.Area = viper.GetInt64("ImageMaxArea")
	ih.Maximums.Width = viper.GetInt("ImageMaxWidth")
	ih.Maximums.Height = viper.GetInt("ImageMaxHeight")

	iiifBase, _ := url.Parse(viper.GetString("IIIFURL"))

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

	// Setup server info in our stats structure
	stats.ServerStart = time.Now()
	stats.RAISVersion = version.Version

	var pubMux = http.NewServeMux()
	handle(pubMux, ih.IIIFBase.Path+"/", http.HandlerFunc(ih.IIIFRoute))
	handle(pubMux, "/images/dzi/", http.HandlerFunc(ih.DZIRoute))

	var srv = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		Addr:         address,
		Handler:      pubMux,
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

		Logger.Infof("RAIS Stopped")
		wait.Done()
	})

	Logger.Infof("RAIS v%s starting...", version.Version)
	serveAsync(&wait, srv)
	wait.Wait()
}

func serveAsync(wait *sync.WaitGroup, srv *http.Server) {
	wait.Add(1)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// Don't report a fatal error when we close the server
			if err != http.ErrServerClosed {
				Logger.Fatalf("Error starting listener: %s", err)
			}
		}
	wait.Done()
	}()
}

func setupCaches() {
	var err error
	icl := viper.GetInt("InfoCacheLen")
	if icl > 0 {
		infoCache, err = lru.New(icl)
		if err != nil {
			Logger.Fatalf("Unable to start info cache: %s", err)
		}
		stats.InfoCache.Enabled = true
	}

	tcl := viper.GetInt("TileCacheLen")
	if tcl > 0 {
		Logger.Debugf("Creating a tile cache to hold up to %d tiles", tcl)
		tileCache, err = lru.New2Q(tcl)
		if err != nil {
			Logger.Fatalf("Unable to start info cache: %s", err)
		}
		stats.TileCache.Enabled = true
	}
}

// handle sends the pattern and raw handler to plugins, and sets up routing on
// whatever is returned (if anything).  All plugins which wrap handlers are
// allowed to run, but the behavior could definitely get weird depending on
// what a given plugin does.  Ye be warned.
func handle(mux *http.ServeMux, pattern string, handler http.Handler) {
	for _, plug := range wrapHandlerPlugins {
		var h2, err = plug(pattern, handler)
		if err == nil {
			handler = h2
		} else if err != plugins.ErrSkipped {
			logger.Fatalf("Error trying to wrap handler %q: %s", pattern, err)
		}
	}
	mux.Handle(pattern, handler)
}
