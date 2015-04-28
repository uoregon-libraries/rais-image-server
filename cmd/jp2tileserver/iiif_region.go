package main

import (
	"strconv"
	"strings"
)

type RegionType int

const (
	RTNone RegionType = iota
	RTFull
	RTPercent
	RTPixel
)

type Region struct {
	Type       RegionType
	X, Y, W, H float64
}

func StringToRegion(p string) Region {
	if p == "full" {
		return Region{Type: RTFull}
	}

	r := Region{Type: RTPixel}
	if len(p) > 4 && p[0:4] == "pct:" {
		r.Type = RTPercent
		p = p[4:]
	}

	vals := strings.Split(p, ",")
	r.X, _ = strconv.ParseFloat(vals[0], 64)
	r.Y, _ = strconv.ParseFloat(vals[1], 64)
	r.W, _ = strconv.ParseFloat(vals[2], 64)
	r.H, _ = strconv.ParseFloat(vals[3], 64)

	return r
}

func (r Region) Valid() bool {
	switch r.Type {
	case RTNone:
		return false
	case RTFull:
		return true
	}

	if r.W <= 0 || r.H <= 0 || r.X < 0 || r.Y < 0 {
		return false
	}

	return true
}
