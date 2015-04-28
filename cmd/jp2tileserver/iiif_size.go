package main

import (
	"strconv"
	"strings"
)

type SizeType int
const (
	STNone SizeType = iota
	STFull
	STScaleToWidth
	STScaleToHeight
	STScalePercent
	STExact
	STBestFit
)
type Size struct {
	Type    SizeType
	Percent float64
	W, H    int
}

func StringToSize(p string) Size {
	if p == "full" {
		return Size{Type: STFull}
	}

	s := Size{Type: STNone}

	if p[0:4] == "pct:" {
		s.Type = STScalePercent
		s.Percent, _ = strconv.ParseFloat(p[4:], 64)
		return s
	}

	if p[0:1] == "!" {
		s.Type = STBestFit
		p = p[1:]
	}

	vals := strings.Split(p, ",")
	s.W, _ = strconv.Atoi(vals[0])
	s.H, _ = strconv.Atoi(vals[1])

	if s.Type == STNone {
		if vals[0] == "" {
			s.Type = STScaleToHeight
		} else if vals[1] == "" {
			s.Type = STScaleToWidth
		} else {
			s.Type = STExact
		}
	}

	return s
}
