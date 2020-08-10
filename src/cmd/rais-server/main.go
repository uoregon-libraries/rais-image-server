package main

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"rais/src/cmd/rais-server/internal/servers"
	"rais/src/iiif"
	"rais/src/img"
	"rais/src/openjpeg"
	"rais/src/plugins"
	"rais/src/version"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
	"github.com/uoregon-libraries/gopkg/interrupts"
	"github.com/uoregon-libraries/gopkg/logger"
)

// Logger is the server's central logger.Logger instance
var Logger *logger.Logger

// Global server stats for admin information gathering
var stats = new(serverStats)

// wait ensures main() doesn't exit until the server(s) are all shutdown
var wait sync.WaitGroup

func main() {
	parseConf()
	Logger = logger.New(logger.LogLevelFromString(viper.GetString("LogLevel")))
	openjpeg.Logger = Logger

	setupCaches()

	var pluginList string

	// Don't let the default plugin list be used if we have an explicit value of ""
	if viper.IsSet("Plugins") {
		pluginList = viper.GetString("Plugins")
	}

	if pluginList == "" || pluginList == "-" {
		Logger.Infof("No plugins will attempt to be loaded")
	} else {
		LoadPlugins(Logger, strings.Split(pluginList, ","))
	}

	// Register our JP2 decoder after plugins have been loaded to allow plugins
	// to handle images - for instance, we might want a pyramidal tiff plugin or
	// something one day
	img.RegisterDecodeHandler(decodeJP2)

	// File streamer for handling images on the local filesystem
	img.RegisterStreamReader(fileStreamReader)

	// Cloud streamer for attempting to handle anything else.  Technically this
	// can do local files, too, but the overhead is just too much if we want to
	// keep showcasing how fast RAIS is with local files....
	img.RegisterStreamReader(cloudStreamReader)

	tilePath := viper.GetString("TilePath")
	webPath := viper.GetString("IIIFWebPath")
	if webPath == "" {
		webPath = "/iiif"
	}
	p2 := path.Clean(webPath)
	if webPath != p2 {
		Logger.Warnf("WebPath %q cleaned; using %q instead", webPath, p2)
		webPath = p2
	}
	address := viper.GetString("Address")
	adminAddress := viper.GetString("AdminAddress")

	ih := NewImageHandler(tilePath, webPath)
	ih.Maximums.Area = viper.GetInt64("ImageMaxArea")
	ih.Maximums.Width = viper.GetInt("ImageMaxWidth")
	ih.Maximums.Height = viper.GetInt("ImageMaxHeight")

	// Check for scheme remapping configuration - if it exists, it's the final id-to-URL handler
	schemeMapConfig := viper.GetString("SchemeMap")
	if schemeMapConfig != "" {
		err := parseSchemeMap(ih, schemeMapConfig)
		if err != nil {
			Logger.Fatalf("Error parsing SchemeMap: %s", err)
		}
	}

	iiifBaseURL := viper.GetString("IIIFBaseURL")
	if iiifBaseURL != "" {
		baseURL, _ := url.Parse(iiifBaseURL)
		Logger.Infof("Explicitly setting IIIF base URL to %q", baseURL)
		ih.BaseURL = baseURL
	}

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
	stats.RAISBuild = version.Build

	// Set up handlers / listeners
	var pubSrv = servers.New("RAIS", address)
	pubSrv.AddMiddleware(logMiddleware)
	handle(pubSrv, ih.WebPathPrefix+"/", http.HandlerFunc(ih.IIIFRoute))
	handle(pubSrv, "/", http.NotFoundHandler())

	var admSrv = servers.New("RAIS Admin", adminAddress)
	admSrv.AddMiddleware(logMiddleware)
	admSrv.HandleExact("/admin/stats.json", stats)
	admSrv.HandlePrefix("/admin/cache/purge", http.HandlerFunc(adminPurgeCache))

	interrupts.TrapIntTerm(shutdown)

	Logger.Infof("RAIS v%s starting...", version.Version)
	servers.ListenAndServe(func(srv *servers.Server, err error) {
		Logger.Errorf("Error running %q server: %s", srv.Name, err)
		shutdown()
	})
	wait.Wait()
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
			Logger.Fatalf("Error trying to wrap handler %q: %s", pattern, err)
		}
	}

	srv.HandlePrefix(pattern, handler)
}

func shutdown() {
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
}

func parseSchemeMap(ih *ImageHandler, schemeMapConfig string) error {
	var confs = strings.Fields(schemeMapConfig)
	for _, conf := range confs {
		var parts = strings.Split(conf, "=")
		if len(parts) != 2 {
			return fmt.Errorf(`invalid scheme map %q: format must be "scheme=prefix"`, conf)
		}
		var scheme, prefix = parts[0], parts[1]

		var err = ih.AddSchemeMap(scheme, prefix)
		if err != nil {
			return fmt.Errorf("invalid scheme map %q: %s", conf, err)
		}
	}

	return nil
}
