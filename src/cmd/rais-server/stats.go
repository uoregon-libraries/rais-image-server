package main

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"
)

type plugStats struct {
	Path      string
	Functions []string
}

type cacheStats struct {
	GetCount   uint64
	GetHits    uint64
	SetCount   uint64
	Length     int
	m          sync.Mutex
	Enabled    bool
	HitPercent float64
}

func (cs *cacheStats) setHitPercent() {
	cs.m.Lock()
	defer cs.m.Unlock()
	if cs.GetCount == 0 {
		cs.HitPercent = 0
		return
	}
	cs.HitPercent = float64(cs.GetHits) / float64(cs.GetCount)
}

// Get increments GetCount safely
func (cs *cacheStats) Get() {
	atomic.AddUint64(&cs.GetCount, 1)
}

// Hit increments GetHits safely
func (cs *cacheStats) Hit() {
	atomic.AddUint64(&cs.GetHits, 1)
}

// Set increments SetCount safely
func (cs *cacheStats) Set() {
	atomic.AddUint64(&cs.SetCount, 1)
}

// serverStats holds a bunch of global data.  This is only threadsafe when
// calling functions, so don't directly manipulate anything except when you
// know only one thread can possibly exist!  (e.g., when first setting up the
// object)
type serverStats struct {
	m           sync.Mutex
	InfoCache   cacheStats
	TileCache   cacheStats
	Plugins     []plugStats
	RAISVersion string
	ServerStart time.Time
	Uptime      string
}

// Serialize writes the stats data to w in JSON format
func (s *serverStats) Serialize() ([]byte, error) {
	s.calculateDerivedStats()
	return json.Marshal(s)
}

// calculateDerivedStats computes things we don't need to store real-time, such
// as cache hit percent, uptime, etc.
func (s *serverStats) calculateDerivedStats() {
	s.m.Lock()

	s.Uptime = time.Since(s.ServerStart).Round(time.Second).String()
	if infoCache != nil {
		s.InfoCache.setHitPercent()
		s.InfoCache.Length = infoCache.Len()
	}
	if tileCache != nil {
		s.TileCache.setHitPercent()
		s.TileCache.Length = tileCache.Len()
	}

	s.m.Unlock()
}
