// cache.go houses all the logic for the various caching built into RAIS as
// well as for sending cache invalidations to plugins

package main

import (
	"rais/src/iiif"

	lru "github.com/hashicorp/golang-lru"
	"github.com/spf13/viper"
)

var infoCache *lru.Cache
var tileCache *lru.TwoQueueCache

// setupCaches looks for config for caching and sets up the tile/info caches
// appropriately.
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

// purgeCaches removes all cached data
func purgeCaches() {
	if tileCache != nil {
		tileCache.Purge()
	}
	if infoCache != nil {
		infoCache.Purge()
	}
}

// expireCachedImage removes cached data for a single IIIF ID.  Unfortunately,
// the tile cache is keyed by the entire IIIF request, not the ID (obviously).
// Since we can't get a list of all cached tiles for a given image, we have to
// purge the whole cache.
func expireCachedImage(id iiif.ID) {
	if tileCache != nil {
		tileCache.Purge()
	}

	if infoCache != nil {
		infoCache.Remove(id)
	}
}
