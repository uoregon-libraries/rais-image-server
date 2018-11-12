package main

import (
	"net/http"
	"rais/src/iiif"
)

func (s *serverStats) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var json, err = s.Serialize()
	if err != nil {
		http.Error(w, "error generating json: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func adminPurgeCache(w http.ResponseWriter, req *http.Request) {
	// All requests must be POST as hitting this endpoint can have serious consequences
	var reqType = req.PostFormValue("type")
	switch reqType {
	case "single":
		var id = iiif.ID(req.PostFormValue("id"))
		expireCachedImage(id)
	case "all":
		purgeCaches()
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Write([]byte("OK"))
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
