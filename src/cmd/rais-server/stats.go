package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type plugStats struct {
	Path      string
	Functions []string
}

type cacheStats struct {
	Enabled    bool
	GetCount   uint64
	GetHits    uint64
	HitPercent float64
	SetCount   uint64
}

func (cs *cacheStats) setHitPercent() {
	if cs.GetCount == 0 {
		cs.HitPercent = 0
		return
	}
	cs.HitPercent = float64(cs.GetHits) / float64(cs.GetCount)
}

type serverStats struct {
	InfoCache   cacheStats
	TileCache   cacheStats
	Plugins     []plugStats
	RAISVersion string
	ServerStart time.Time
	Uptime      string
}

func (s *serverStats) setUptime() {
	s.Uptime = time.Since(s.ServerStart).Round(time.Second).String()
}

// Serialize writes the stats data to w in JSON format
func (s *serverStats) Serialize() ([]byte, error) {
	// Calculate derived stats only on serializations
	s.setUptime()
	s.InfoCache.setHitPercent()
	s.TileCache.setHitPercent()

	return json.Marshal(s)
}

func (s *serverStats) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var json, err = s.Serialize()
	if err != nil {
		http.Error(w, "error generating json: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
