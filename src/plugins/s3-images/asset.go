package main

import (
	"hash/fnv"
	"path/filepath"
	"rais/src/iiif"
	"strconv"
	"sync"
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
	id   iiif.ID
	key  string
	path string
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
