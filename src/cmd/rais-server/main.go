package main

import (
	"fmt"
	"iiif"
	"net/http"
	"net/url"
	"openjpeg"
	"os"
	"time"
	"version"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/golang-lru"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	l "github.com/uoregon-libraries/gopkg/logger"
)

var tilePath string
var infoCache *lru.Cache
var tileCache *lru.TwoQueueCache
var logger *l.Logger

const defaultAddress = ":12415"
const defaultInfoCacheLen = 10000

var defaultLogLevel = l.Debug.String()

func main() {
	// Defaults
	viper.SetDefault("Address", defaultAddress)
	viper.SetDefault("InfoCacheLen", defaultInfoCacheLen)
	viper.SetDefault("LogLevel", defaultLogLevel)

	// Allow all configuration to be in environment variables
	viper.SetEnvPrefix("RAIS")
	viper.AutomaticEnv()

	// Config file options
	viper.SetConfigName("rais")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	// CLI flags
	pflag.String("iiif-url", "", `Base URL for serving IIIF requests, e.g., "http://example.com:8888/images/iiif"`)
	viper.BindPFlag("IIIFURL", pflag.CommandLine.Lookup("iiif-url"))
	pflag.String("address", defaultAddress, "http service address")
	viper.BindPFlag("Address", pflag.CommandLine.Lookup("address"))
	pflag.String("tile-path", "", "Base path for images")
	viper.BindPFlag("TilePath", pflag.CommandLine.Lookup("tile-path"))
	pflag.Int("iiif-info-cache-size", defaultInfoCacheLen, "Maximum cached image info entries (IIIF only)")
	viper.BindPFlag("InfoCacheLen", pflag.CommandLine.Lookup("iiif-info-cache-size"))
	pflag.String("capabilities-file", "", "TOML file describing capabilities, rather than everything RAIS supports")
	viper.BindPFlag("CapabilitiesFile", pflag.CommandLine.Lookup("capabilities-file"))
	pflag.String("log-level", defaultLogLevel, "Log level: the server will only log notifications at "+
		"this level and above (must be DEBUG, INFO, WARN, ERROR, or CRIT)")
	viper.BindPFlag("LogLevel", pflag.CommandLine.Lookup("log-level"))

	pflag.Parse()

	// Make sure required values exist
	if !viper.IsSet("TilePath") {
		fmt.Println("ERROR: --tile-path is required")
		pflag.Usage()
		os.Exit(1)
	}

	// Make sure we have a valid log level
	var level = l.LogLevelFromString(viper.GetString("LogLevel"))
	if level == l.Invalid {
		fmt.Println("ERROR: --log-level must be DEBUG, INFO, WARN, ERROR, or CRIT")
		pflag.Usage()
		os.Exit(1)
	}
	logger = l.New(level)
	openjpeg.Logger = logger

	// Pull all values we need for all cases
	tilePath = viper.GetString("TilePath")
	address := viper.GetString("Address")

	// Handle IIIF data only if we have a IIIF URL
	ih := NewImageHandler(tilePath)
	if viper.IsSet("IIIFURL") {
		logger.Debugf("Attempting to start up IIIF at %s", viper.GetString("IIIFURL"))
		iiifURL := viper.GetString("IIIFURL")
		iiifBase, err := url.Parse(iiifURL)
		if err == nil && iiifBase.Scheme == "" {
			err = fmt.Errorf("empty scheme")
		}
		if err == nil && iiifBase.Host == "" {
			err = fmt.Errorf("empty host")
		}
		if err == nil && iiifBase.Path == "" {
			err = fmt.Errorf("empty path")
		}
		if err != nil {
			logger.Fatalf("Invalid IIIF URL (%s) specified: %s", iiifURL, err)
		}

		icl := viper.GetInt("InfoCacheLen")
		if icl > 0 {
			infoCache, err = lru.New(icl)
			if err != nil {
				logger.Fatalf("Unable to start info cache: %s", err)
			}
		}

		tcl := viper.GetInt("TileCacheLen")
		if tcl > 0 {
			logger.Debugf("Creating a tile cache to hold up to %d tiles", tcl)
			tileCache, err = lru.New2Q(tcl)
			if err != nil {
				logger.Fatalf("Unable to start info cache: %s", err)
			}
		}

		logger.Infof("IIIF enabled at %s", iiifBase.String())
		ih.EnableIIIF(iiifBase)

		if viper.IsSet("CapabilitiesFile") {
			filename := viper.GetString("CapabilitiesFile")
			ih.FeatureSet = &iiif.FeatureSet{}
			_, err := toml.DecodeFile(filename, &ih.FeatureSet)
			if err != nil {
				logger.Fatalf("Invalid file or formatting in capabilities file '%s'", filename)
			}
			logger.Debugf("Setting IIIF capabilities from file '%s'", filename)
		}

		http.HandleFunc(ih.IIIFBase.Path+"/", ih.IIIFRoute)
		http.HandleFunc("/images/dzi/", ih.DZIRoute)
	}

	http.HandleFunc("/images/tiles/", TileHandler)
	http.HandleFunc("/images/resize/", ResizeHandler)
	http.HandleFunc("/version", VersionHandler)

	logger.Infof("RAIS v%s starting...", version.Version)
	var srv = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		Addr:         address,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("Error starting listener: %s", err)
	}
}

// VersionHandler spits out the raw version string to the browser
func VersionHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(version.Version))
}
