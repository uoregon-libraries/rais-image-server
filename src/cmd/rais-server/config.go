package main

import (
	"fmt"
	"math"
	"net/url"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/uoregon-libraries/gopkg/logger"
)

// parseConf centralizes all config reading and validating for the core RAIS options
func parseConf() {
	// Default configuration values
	var defaultAddress = ":12415"
	var defaultAdminAddress = ":12416"
	var defaultInfoCacheLen = 10000
	var defaultLogLevel = logger.Debug.String()
	var defaultPlugins = "s3-images.so,json-tracer.so"

	// Defaults
	viper.SetDefault("Address", defaultAddress)
	viper.SetDefault("AdminAddress", defaultAdminAddress)
	viper.SetDefault("InfoCacheLen", defaultInfoCacheLen)
	viper.SetDefault("LogLevel", defaultLogLevel)
	viper.SetDefault("Plugins", defaultPlugins)

	// Allow all configuration to be in environment variables
	viper.SetEnvPrefix("RAIS")
	viper.AutomaticEnv()

	// Config file options
	viper.SetConfigName("rais")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("ERROR: Invalid RAIS config file (/etc/rais.toml or ./rais.toml): %s\n", err)
		os.Exit(1)
	}

	// CLI flags
	pflag.String("iiif-url", "", `Base URL for serving IIIF requests, e.g., "http://example.com/images/iiif"`)
	viper.BindPFlag("IIIFURL", pflag.CommandLine.Lookup("iiif-url"))
	pflag.String("address", defaultAddress, "http service address")
	viper.BindPFlag("Address", pflag.CommandLine.Lookup("address"))
	pflag.String("admin-address", defaultAdminAddress, "http service for administrative endpoints")
	viper.BindPFlag("AdminAddress", pflag.CommandLine.Lookup("admin-address"))
	pflag.String("tile-path", "", "Base path for images")
	viper.BindPFlag("TilePath", pflag.CommandLine.Lookup("tile-path"))
	pflag.Int("iiif-info-cache-size", defaultInfoCacheLen, "Maximum cached image info entries (IIIF only)")
	viper.BindPFlag("InfoCacheLen", pflag.CommandLine.Lookup("iiif-info-cache-size"))
	pflag.String("capabilities-file", "", "TOML file describing capabilities, rather than everything RAIS supports")
	viper.BindPFlag("CapabilitiesFile", pflag.CommandLine.Lookup("capabilities-file"))
	pflag.String("log-level", defaultLogLevel, "Log level: the server will only log notifications at "+
		"this level and above (must be DEBUG, INFO, WARN, ERROR, or CRIT)")
	viper.BindPFlag("LogLevel", pflag.CommandLine.Lookup("log-level"))
	pflag.Int64("image-max-area", math.MaxInt64, "Maximum area (w x h) of images to be served")
	viper.BindPFlag("ImageMaxArea", pflag.CommandLine.Lookup("image-max-area"))
	pflag.Int("image-max-width", math.MaxInt32, "Maximum width of images to be served")
	viper.BindPFlag("ImageMaxWidth", pflag.CommandLine.Lookup("image-max-width"))
	pflag.Int("image-max-height", math.MaxInt32, "Maximum height of images to be served")
	viper.BindPFlag("ImageMaxHeight", pflag.CommandLine.Lookup("image-max-height"))
	pflag.String("plugins", defaultPlugins, "comma-separated plugin pattern list, e.g., "+
		`"s3-images.so,datadog.so,json-tracer.so,/opt/rais/plugins/*.so"`)
	viper.BindPFlag("Plugins", pflag.CommandLine.Lookup("plugins"))

	pflag.Parse()

	// Make sure required values exist
	if !viper.IsSet("TilePath") {
		fmt.Println("ERROR: tile path is required")
		pflag.Usage()
		os.Exit(1)
	}

	var level = logger.LogLevelFromString(viper.GetString("LogLevel"))
	if level == logger.Invalid {
		fmt.Println("ERROR: Invalid log level (must be DEBUG, INFO, WARN, ERROR, or CRIT)")
		pflag.Usage()
		os.Exit(1)
	}

	var iiifURL = viper.GetString("IIIFURL")
	if iiifURL == "" {
		fmt.Println("ERROR: IIIF URL may not be blank")
		pflag.Usage()
		os.Exit(1)
	}
	var u, err = url.Parse(iiifURL)
	if err == nil && u.Scheme == "" {
		err = fmt.Errorf("empty scheme")
	}
	if err == nil && u.Host == "" {
		err = fmt.Errorf("empty host")
	}
	if err == nil && u.Path == "" {
		err = fmt.Errorf("empty path")
	}
	if err != nil {
		fmt.Printf("ERROR: invalid IIIF URL (%s) specified: %s\n", iiifURL, err)
	}
}
