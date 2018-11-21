package main

import (
	"math/rand"
	"rais/src/iiif"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/uoregon-libraries/gopkg/assert"
)

func TestAssetLookup(t *testing.T) {
	s3cache = "/tmp"
	t.Run("S3 ID", func(t *testing.T) {
		var a = lookupAsset(iiif.ID("s3:foo"))
		assert.Equal("foo", a.key, "key", t)
		assert.Equal("/tmp/13/83/foo", a.path, "path", t)
		assert.Equal(iiif.ID("s3:foo"), a.id, "id", t)
	})
	t.Run("non-S3 ID", func(t *testing.T) {
		var a = lookupAsset(iiif.ID("foo"))
		assert.Equal("", a.key, "empty key", t)
	})
	t.Run("existing ID", func(t *testing.T) {
		assets = make(map[iiif.ID]*asset)
		var a = lookupAsset(iiif.ID("s3:foo"))
		var b = lookupAsset(iiif.ID("s3:foo"))
		assert.Equal(a, b, "same asset", t)
		assert.Equal(1, len(assets), "len(assets)", t)
	})
}

func TestFLock(t *testing.T) {
	var a = lookupAsset(iiif.ID("s3:foo"))

	// Set up intense concurrency to see if we can cause mayhem
	var successes uint32
	var wg sync.WaitGroup
	var tryit = func() {
		time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(10)))
		if a.tryFLock() {
			atomic.AddUint32(&successes, 1)
		}
		wg.Done()
	}
	for x := 0; x < 100; x++ {
		wg.Add(1)
		go tryit()
	}
	wg.Wait()

	assert.Equal(uint32(1), successes, "only one tryFLock call succeeds", t)
	a.fUnlock()
	assert.True(a.tryFLock(), "tryFLock call succeeds after fUnlock", t)
	a.fUnlock()
}
