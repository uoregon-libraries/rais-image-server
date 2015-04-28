package main

import (
	"testing"
)

func assert(expression bool, message string, t *testing.T) {
	if !expression {
		t.Errorf(message)
		return
	}
	t.Log(message)
}

func assertEqual(expected, actual interface{}, message string, t *testing.T) {
	if expected != actual {
		t.Errorf("Expected %#v, but got %#v - %s", expected, actual, message)
		return
	}
	t.Log(message)
}

func TestRegionTypePercent(t *testing.T) {
	r := StringToRegion("pct:41.6,7.5,40,70")
	assert(r.Type == RTPercent, "r.Type == RTPercent", t)
	assertEqual(41.6, r.X, "r.X", t)
	assertEqual(7.5, r.Y, "r.Y", t)
	assertEqual(40.0, r.W, "r.W", t)
	assertEqual(70.0, r.H, "r.H", t)
}

func TestRegionTypePixels(t *testing.T) {
	r := StringToRegion("10,10,40,70")
	assert(r.Type == RTPixel, "r.Type == RTPixel", t)
	assertEqual(10.0, r.X, "r.X", t)
	assertEqual(10.0, r.Y, "r.Y", t)
	assertEqual(40.0, r.W, "r.W", t)
	assertEqual(70.0, r.H, "r.H", t)
}

func TestRegionTypeFull(t *testing.T) {
	r := StringToRegion("full")
	assert(r.Type == RTFull, "r.Type == RTFull", t)
}
