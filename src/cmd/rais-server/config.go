package main

import (
	"fmt"
	"math"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// parseConf centralizes all config reading and validating for the core RAIS options
func parseConf() {
	// Defaults
	viper.SetDefault("Address", defaultAddress)
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
}
