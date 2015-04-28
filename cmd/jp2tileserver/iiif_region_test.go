package main

import (
	"fmt"
	"testing"
	"regexp"
	"runtime"
)

var re = regexp.MustCompile(`^.*jp2tileserver\.(.*)$`)

type Caller struct {
	Func     *runtime.Func
	Name     string
	Filename string
	Line     int
}

func getCallerName(skip int) *Caller {
	// Increase skip since they surely don't want *this* function
	pc, file, line, _ := runtime.Caller(skip + 1)
	fn := runtime.FuncForPC(pc)
	return &Caller{
		Func:     fn,
		Name:     re.ReplaceAllString(fn.Name(), "$1"),
		Filename: file,
		Line:     line,
	}
}

func success(caller *Caller, message string, t *testing.T) {
	fmt.Printf("\033[32mok\033[0m        %s(): %s\n", caller.Name, message)
}

func failure(caller *Caller, message string, t *testing.T) {
	fmt.Printf("\033[31;1mnot ok\033[0m    %s(): %s\n", caller.Name, message)
	fmt.Printf("          - %s:%d\n", caller.Filename, caller.Line)
	t.Fail()
}

func assert(expression bool, message string, t *testing.T) {
	caller := getCallerName(1)
	if !expression {
		failure(caller, message, t)
		return
	}
	success(caller, message, t)
}

func assertEqual(expected, actual interface{}, message string, t *testing.T) {
	caller := getCallerName(1)
	if expected != actual {
		failure(caller, fmt.Sprintf("Expected %#v, but got %#v - %s", expected, actual, message), t)
		return
	}
	success(caller, message, t)
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
