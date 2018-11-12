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
	m          sync.Mutex
	Enabled    bool
	GetCount   uint64
	GetHits    uint64
	HitPercent float64
	SetCount   uint64
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
	RAISBuild   string
	ServerStart time.Time
	Uptime      string
}

func (s *serverStats) setUptime() {
	s.m.Lock()
	s.Uptime = time.Since(s.ServerStart).Round(time.Second).String()
	s.m.Unlock()
}

// Serialize writes the stats data to w in JSON format
func (s *serverStats) Serialize() ([]byte, error) {
	// Calculate derived stats only on serializations
	s.setUptime()
	s.InfoCache.setHitPercent()
	s.TileCache.setHitPercent()

	return json.Marshal(s)
}
