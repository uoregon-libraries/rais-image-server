package main

import (
	"net/http"
	"net/url"
	"rais/src/cmd/rais-server/internal/servers"
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
	adminAddress := viper.GetString("AdminAddress")

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

	// Set up handlers / listeners
	var pubSrv = servers.New("RAIS", address)
	handle(pubSrv, ih.IIIFBase.Path+"/", http.HandlerFunc(ih.IIIFRoute))
	handle(pubSrv, "/images/dzi/", http.HandlerFunc(ih.DZIRoute))
	var admSrv = servers.New("RAIS Admin", adminAddress)
	admSrv.Handle("/admin/stats.json", stats)

	var wait sync.WaitGroup
	interrupts.TrapIntTerm(func() {
		wait.Add(1)
		Logger.Infof("Stopping RAIS...")
		servers.Shutdown(nil)

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
	servers.ListenAndServe(func(srv *servers.Server, err error) {
		Logger.Errorf("Error running %q server: %s", srv.Name, err)
	})
	wait.Wait()
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
func handle(srv *servers.Server, pattern string, handler http.Handler) {
	for _, plug := range wrapHandlerPlugins {
		var h2, err = plug(pattern, handler)
		if err == nil {
			handler = h2
		} else if err != plugins.ErrSkipped {
			logger.Fatalf("Error trying to wrap handler %q: %s", pattern, err)
		}
	}
	srv.Handle(pattern, handler)
}
