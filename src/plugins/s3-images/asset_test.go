package main

import (
	"rais/src/iiif"
	"testing"

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
	s3cache = "/tmp"
	var a = lookupAsset(iiif.ID("s3:foo"))
	assert.True(a.tryFLock(), "first tryFlock call succeeds", t)
	assert.False(a.tryFLock(), "second tryFlock call fails", t)
}
