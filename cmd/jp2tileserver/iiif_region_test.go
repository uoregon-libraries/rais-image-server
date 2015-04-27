package main

import (
	"testing"
)

func assertEqualF(expected, actual float64, message string, t *testing.T) {
	if expected != actual {
		t.Errorf("Expected %g, but got %g - %s", expected, actual, message)
		return
	}
	t.Log(message)
}

func TestPercent(t *testing.T) {
	r := StringToRegion("pct:41.6,7.5,40,70")
	if r.Type != RTPercent {
		t.Errorf("Expected r.Type to be RTPercent")
	}

	assertEqualF(41.6, r.X, "r.X", t)
	assertEqualF(7.5, r.Y, "r.Y", t)
	assertEqualF(40, r.W, "r.W", t)
	assertEqualF(70, r.H, "r.H", t)
}

func TestPixels(t *testing.T) {
	r := StringToRegion("10,10,40,70")
	if r.Type != RTPixel {
		t.Errorf("Expected r.Type to be RTPixel")
	}

	assertEqualF(10, r.X, "r.X", t)
	assertEqualF(10, r.Y, "r.Y", t)
	assertEqualF(40, r.W, "r.W", t)
	assertEqualF(70, r.H, "r.H", t)
}

func TestFull(t *testing.T) {
	r := StringToRegion("full")
	if r.Type != RTFull {
		t.Errorf("Expected r.Type to be RTFull")
	}
}
