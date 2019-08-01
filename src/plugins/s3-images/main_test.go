package main

import (
	"rais/src/iiif"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

func TestIDToPath(t *testing.T) {
	// Set up an asset hard-coded to /dev/null so the fetch logic is skipped -
	// download will stat the file, be tricked into thnking it's cached, and return
	var a, _ = lookupAsset(iiif.ID("nil://fakebucket/fakeid"))
	a.path = "/dev/null"

	// Set up intense concurrency to see if we can cause mayhem
	var wg sync.WaitGroup
	var successes uint32
	var tryit = func() {
		defer wg.Done()

		var path, err = IDToPath("nil://fakebucket/fakeid")
		if err != nil {
			t.Errorf("Failed trying to get path from ID: %s", err)
			t.FailNow()
		}
		if path != "/dev/null" {
			t.Errorf("Unexpected path from IDToPath: %q", path)
			t.FailNow()
		}
		atomic.AddUint32(&successes, 1)
	}
	for x := 0; x < 100; x++ {
		wg.Add(1)
		go tryit()
	}
	wg.Wait()

	assert.Equal(uint32(100), successes, "all IDToPath calls worked", t)
}
