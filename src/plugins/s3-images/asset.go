package main

import (
	"hash/fnv"
	"os"
	"path/filepath"
	"rais/src/iiif"
	"strconv"
	"sync"
	"sync/atomic"
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
	inUse      uint32
	fs         sync.Mutex
	lockreader sync.Mutex
}

func lookupAsset(id iiif.ID) *asset {
	assetMutex.Lock()
	var a = assets[id]
	if a == nil {
		a = &asset{id: id, key: makeKey(id)}
		a.path = makePath(a.key)
		assets[id] = a
	}
	assetMutex.Unlock()

	return a
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
	var inUse = atomic.LoadUint32(&a.inUse) == 1
	if !inUse {
		a.fs.Lock()
		atomic.StoreUint32(&a.inUse, 1)
	}
	a.lockreader.Unlock()

	return !inUse
}

func (a *asset) fUnlock() {
	a.lockreader.Lock()
	atomic.StoreUint32(&a.inUse, 0)
	a.fs.Unlock()
	a.lockreader.Unlock()
}
