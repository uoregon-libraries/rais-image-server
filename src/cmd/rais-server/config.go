package main

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"path"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/uoregon-libraries/gopkg/logger"
)

// cleanWebPath normalizes a configured IIIF web path.  An empty string is left
// empty (which disables that endpoint); anything else is run through path.Clean.
func cleanWebPath(p string) string {
	if p == "" {
		return ""
	}
	return path.Clean(p)
}

// parseConf centralizes all config reading and validating for the core RAIS options
func parseConf() {
	// Default configuration values
	var defaultAddress = ":12415"
	var defaultAdminAddress = ":12416"
	var defaultInfoCacheLen = 10000
	var defaultLogLevel = logger.Debug.String()
	var defaultPlugins = "-"
	var defaultJPGQuality = 75

	// Defaults
	viper.SetDefault("Address", defaultAddress)
	viper.SetDefault("AdminAddress", defaultAdminAddress)
	viper.SetDefault("InfoCacheLen", defaultInfoCacheLen)
	viper.SetDefault("LogLevel", defaultLogLevel)
	viper.SetDefault("Plugins", defaultPlugins)
	viper.SetDefault("JPGQuality", defaultJPGQuality)

	// Allow all configuration to be in environment variables.  AllowEmptyEnv lets
	// an explicitly-empty env var (e.g., RAIS_IIIFWEBPATHV3="") override a default,
	// which is how a spec version's endpoint is disabled via the environment.
	viper.SetEnvPrefix("RAIS")
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()

	// Config file options
	viper.SetConfigName("rais")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("ERROR: Invalid RAIS config file (/etc/rais.toml or ./rais.toml): %s\n", err)
			os.Exit(1)
		}
	}

	// CLI flags
	pflag.String("iiif-base-url", "", "Base URL for RAIS to report in info.json requests "+
		"(defaults to the requests as they come in, so you probably don't want to set this)")
	viper.BindPFlag("IIIFBaseURL", pflag.CommandLine.Lookup("iiif-base-url"))
	pflag.String("iiif-web-path-v2", "/iiif/v2", `Base path for serving IIIF 2.1 requests, e.g., "/iiif/v2" (empty disables v2)`)
	viper.BindPFlag("IIIFWebPathV2", pflag.CommandLine.Lookup("iiif-web-path-v2"))
	pflag.String("iiif-web-path-v3", "/iiif/v3", `Base path for serving IIIF 3.0 requests, e.g., "/iiif/v3" (empty disables v3)`)
	viper.BindPFlag("IIIFWebPathV3", pflag.CommandLine.Lookup("iiif-web-path-v3"))
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
		`"json-tracer.so,/opt/rais/plugins/*.so"`)
	viper.BindPFlag("Plugins", pflag.CommandLine.Lookup("plugins"))
	pflag.Int("jpg-quality", 75, "Quality of JPEG output")
	viper.BindPFlag("JPGQuality", pflag.CommandLine.Lookup("jpg-quality"))
	pflag.String("scheme-map", "", "Whitespace-delimited map of scheme to prefix, e.g., "+
		`"acme=s3://bucket1 marc=s3://bucket2/some/path"`)
	viper.BindPFlag("SchemeMap", pflag.CommandLine.Lookup("scheme-map"))

	pflag.Parse()

	// Make sure required values exist
	if viper.GetString("TilePath") == "" {
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

	// Friendly migration for the old single "IIIFWebPath" setting: if it's still
	// set (config file or RAIS_IIIFWEBPATH) and neither replacement has been set
	// explicitly, honor it as the v2 path and disable v3 so existing deployments
	// keep serving the exact URLs they did before upgrading.
	if viper.IsSet("IIIFWebPath") && !viper.IsSet("IIIFWebPathV2") && !viper.IsSet("IIIFWebPathV3") {
		var oldPath = viper.GetString("IIIFWebPath")
		if oldPath == "" {
			// The old setting treated empty as the default path
			oldPath = "/iiif"
		}
		fmt.Printf("WARNING: IIIFWebPath has been replaced by IIIFWebPathV2 and IIIFWebPathV3; "+
			"serving IIIF 2.1 on %q and disabling IIIF 3.0 support.  Set IIIFWebPathV2 to silence "+
			"this warning, and IIIFWebPathV3 to enable IIIF 3.0.\n", oldPath)
		viper.Set("IIIFWebPathV2", oldPath)
		viper.Set("IIIFWebPathV3", "")
	}

	// Validate the two IIIF web paths.  Either (but not both) may be empty to
	// disable that spec version, and the two cannot resolve to the same path.
	var webPathV2 = cleanWebPath(viper.GetString("IIIFWebPathV2"))
	var webPathV3 = cleanWebPath(viper.GetString("IIIFWebPathV3"))
	if webPathV2 == "" && webPathV3 == "" {
		fmt.Println("ERROR: at least one of IIIFWebPathV2 or IIIFWebPathV3 must be set")
		pflag.Usage()
		os.Exit(1)
	}
	if webPathV2 != "" && webPathV2 == webPathV3 {
		fmt.Printf("ERROR: IIIFWebPathV2 and IIIFWebPathV3 cannot be the same path (%q)\n", webPathV2)
		pflag.Usage()
		os.Exit(1)
	}

	var baseIIIFURL = viper.GetString("IIIFBaseURL")
	if baseIIIFURL != "" {
		var u, err = url.Parse(baseIIIFURL)
		if err == nil && u.Scheme == "" {
			err = fmt.Errorf("empty scheme")
		}
		if err == nil && u.Host == "" {
			err = fmt.Errorf("empty host")
		}
		if err == nil && u.Path != "" {
			err = fmt.Errorf("only scheme and hostname may be specified")
		}
		if err != nil {
			fmt.Printf("ERROR: invalid Base IIIF URL (%s) specified: %s\n", baseIIIFURL, err)
			pflag.Usage()
			os.Exit(1)
		}
	}
}
