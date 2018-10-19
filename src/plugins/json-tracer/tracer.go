package main

import (
	"net/http"
	"rais/src/iiif"
	"strings"
	"sync"
	"time"
)

type trace struct {
	Path     string
	Type     string
	Start    time.Time
	Duration float64
	Status   int
}

type traceList struct {
	createdAt time.Time
	list      []trace
}

func newTraceList() *traceList {
	return &traceList{list: make([]trace, 0, 256)}
}

type tracer struct {
	sync.Mutex
	done      bool
	handler   http.Handler
	traceList *traceList
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// ServeHTTP implements http.Handler.  We call the underlying handler and store
// timing data locally.
func (t *tracer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var now = time.Now()
	var sr = statusRecorder{w, 200}
	t.handler.ServeHTTP(&sr, req)
	var path = req.URL.RawPath
	if path == "" {
		path = req.URL.Path
	}
	t.appendTrace(path, now, time.Since(now), sr.status)
}

// getReqType is a bit ugly and hacky, but attempts to determine what kind of
// IIIF request we have, if any.  Determining type isn't as easy without
// peeking at RAIS's config, which seems like it could break one day (say the
// IIIF URL config changes but we forget to update this plugin).  Instead, we
// just do what we can by pulling the URL apart as necessary to make pretty
// good guesses.
func getReqType(path string) string {
	if len(path) < 9 {
		return "None"
	}

	if path[len(path)-9:] == "info.json" {
		return "Info"
	}

	var parts = strings.Split(path, "/")
	if len(parts) < 5 {
		return "None"
	}

	var iiifPath = strings.Join(parts[len(parts)-5:], "/")
	var u, err = iiif.NewURL(iiifPath)
	if err != nil || !u.Valid() {
		return "None"
	}

	if err == nil {
		if u.Region.Type == iiif.RTFull || u.Region.Type == iiif.RTSquare {
			return "Resize"
		} else if u.Size.W <= 1024 && u.Size.H <= 1024 {
			return "Tile"
		}
	}

	return "Unknown"
}

func (t *tracer) appendTrace(path string, start time.Time, duration time.Duration, status int) {
	t.Lock()
	defer t.Unlock()

	if len(t.traceList.list) == 0 {
		t.traceList.createdAt = time.Now()
	}
	t.traceList.list = append(t.traceList.list, trace{
		Path:     path,
		Type:     getReqType(path),
		Start:    start,
		Duration: duration.Seconds(),
		Status:   status,
	})
}

// loop checks regularly for the last flush having been long enough ago to
// flush to disk again.  This must run in a background goroutine.
func (t *tracer) loop() {
	for {
		var pending []trace
		t.Lock()
		var tlen = len(t.traceList.list)
		if tlen > 0 && time.Since(t.traceList.createdAt) > flushTime {
			pending = t.traceList.list
			t.traceList = newTraceList()
		}
		var done = t.done
		t.Unlock()

		if len(pending) > 0 {
			writeTraces(pending)
		}

		if done {
			return
		}

		time.Sleep(5 * time.Second)
	}
}

func (t *tracer) shutdown(wg *sync.WaitGroup) {
	t.Lock()
	writeTraces(t.traceList.list)
	t.done = true
	t.Unlock()
	wg.Done()
}

type registry struct {
	sync.Mutex
	list []*tracer
}

func (r *registry) new(h http.Handler) *tracer {
	var t = &tracer{handler: h, traceList: newTraceList()}
	go t.loop()
	r.Lock()
	r.list = append(r.list, t)
	r.Unlock()
	return t
}

func (r *registry) shutdown() {
	r.Lock()
	defer r.Unlock()

	var wg sync.WaitGroup
	for _, t := range r.list {
		wg.Add(1)
		go t.shutdown(&wg)
	}

	wg.Wait()
}
