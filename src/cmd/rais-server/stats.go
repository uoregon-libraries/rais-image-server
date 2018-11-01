package main

import (
	"encoding/json"
	"io"
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
	cs.HitPercent = float64(cs.GetHits) / float64(cs.GetCount)
}

type serverStats struct {
	InfoCache   cacheStats
	TileCache   cacheStats
	Plugins     []plugStats
	RAISVersion string
	ServerStart time.Time
	Uptime      time.Duration
}

func (s *serverStats) setUptime() {
	s.Uptime = time.Now().Sub(s.ServerStart)
}

// Serialize writes the stats data to w in JSON format
func (s *serverStats) Serialize(w io.Writer) error {
	// Calculate derived stats only on serializations
	s.setUptime()
	s.InfoCache.setHitPercent()
	s.TileCache.setHitPercent()

	var b, err = json.Marshal(s)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	return err
}
