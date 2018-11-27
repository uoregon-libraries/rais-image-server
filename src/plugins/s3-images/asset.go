package main

import (
	"hash/fnv"
	"os"
	"path/filepath"
	"rais/src/iiif"
	"strconv"
	"sync"
	"time"
)

var assets = make(map[iiif.ID]*asset)
var assetMutex sync.Mutex

func makeKey(id iiif.ID) string {
	var s = string(id)
	if len(s) < 4 || s[:3] != "s3:" {
		return ""
	}
	return s[3:]
}

func buckets(s3ID string) (string, string) {
	var h = fnv.New32()
	h.Write([]byte(s3ID))
	var val = int(h.Sum32() / 10000)
	return strconv.Itoa(val % 100), strconv.Itoa((val / 100) % 100)
}

func makePath(key string) string {
	var bucket1, bucket2 = buckets(key)
	return filepath.Join(s3cache, bucket1, bucket2, key)
}

type asset struct {
	id         iiif.ID
	key        string
	path       string
	inUse      bool
	fs         sync.Mutex
	lockreader sync.Mutex
	lastAccess    time.Time
}

func lookupAsset(id iiif.ID) (a *asset, ok bool) {
	assetMutex.Lock()
	a, ok = assets[id]
	if !ok {
		a = &asset{id: id, key: makeKey(id)}
		a.path = makePath(a.key)
		assets[id] = a
	}
	assetMutex.Unlock()

	return a, ok
}

func (a *asset) s3Get() error {
	// If the file has already been cached, we can just return here
	var _, err = os.Stat(a.path)
	if err == nil {
		return nil
	}

	l.Debugf("s3-images plugin: no cached file at %q; downloading from S3", a.path)
	return a.fetch()
}

// tryFLock attempts to lock for file writing in a non-blocking way.  If the
// lock can be acquired, the return is true, otherwise false.
func (a *asset) tryFLock() bool {
	a.lockreader.Lock()
	var inUse = a.inUse
	if !inUse {
		a.fs.Lock()
		a.inUse = true
	}
	a.lockreader.Unlock()

	return !inUse
}

// For when master Yoda's around.  There is no try.
func (a *asset) fLock() {
	a.lockreader.Lock()
	a.fs.Lock()
	a.inUse = true
	a.lockreader.Unlock()
}

func (a *asset) fUnlock() {
	a.lockreader.Lock()
	a.inUse = false
	a.fs.Unlock()
	a.lockreader.Unlock()
}

// read lets us track when an asset is being requested.  For the moment we just
// track a timestamp, but we could also track other stats to improve how we
// decide what to purge from the local filesystem.
func (a *asset) read() {
	a.lastAccess = time.Now().Add(cacheLifetime)
}

// purge locks the asset, deletes it from the filesystem, and untracks it from
// the assets list.  This doesn't return an error, instead logging inline if
// the asset can't be deleted, because we are calling this asynchronously to
// avoid potentially long delays if the asset is mid-download right when it's
// being purged.
func (a *asset) purge() {
	var err = os.Remove(a.path)
	if err != nil && !os.IsNotExist(err) {
		l.Errorf("s3-images plugin: Unable to purge cached file at %q: %s", a.path, err)
		return
	}
}
