package main

import (
	"testing"
)

func TestSizeTypeFull(t *testing.T) {
	s := StringToSize("full")
	assertEqual(STFull, s.Type, "s.Type == STFull", t)
}

func TestSizeTypeScaleWidth(t *testing.T) {
	s := StringToSize("125,")
	assertEqual(STScaleToWidth, s.Type, "s.Type == STScaleToWidth", t)
	assertEqual(125, s.W, "s.W", t)
	assertEqual(0, s.H, "s.H", t)
}

func TestSizeTypeScaleHeight(t *testing.T) {
	s := StringToSize(",250")
	assertEqual(STScaleToHeight, s.Type, "s.Type == STScaleToHeight", t)
	assertEqual(0, s.W, "s.W", t)
	assertEqual(250, s.H, "s.H", t)
}

func TestSizeTypePercent(t *testing.T) {
	s := StringToSize("pct:41.6")
	assertEqual(STScalePercent, s.Type, "s.Type == STScalePercent", t)
	assertEqual(41.6, s.Percent, "s.Percent", t)
}

func TestSizeTypeExact(t *testing.T) {
	s := StringToSize("125,250")
	assertEqual(STExact, s.Type, "s.Type == STExact", t)
	assertEqual(125, s.W, "s.W", t)
	assertEqual(250, s.H, "s.H", t)
}

func TestSizeTypeBestFit(t *testing.T) {
	s := StringToSize("!25,50")
	assertEqual(STBestFit, s.Type, "s.Type == STBestFit", t)
	assertEqual(25, s.W, "s.W", t)
	assertEqual(50, s.H, "s.H", t)
}
