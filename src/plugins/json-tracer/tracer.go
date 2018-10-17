package main

import (
	"net/http"
	"sync"
	"time"
)

type trace struct {
	Path     string
	Start    time.Time
	Duration float64
	Status   int
}

type traceList struct {
	createdAt time.Time
	list      []trace
}

func newTraceList() *traceList {
	return &traceList{
		createdAt: time.Now(),
		list:      make([]trace, 0, maxTraces),
	}
}

type tracer struct {
	sync.Mutex
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
// timing data locally.  If we have enough timing data, we send it off to be
// written to disk.
func (t *tracer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var now = time.Now()
	var sr = statusRecorder{w, 200}
	t.handler.ServeHTTP(&sr, req)
	var path = req.URL.RawPath
	if path == "" {
		path = req.URL.Path
	}
	go t.appendTrace(path, now, time.Since(now), sr.status)
}

func (t *tracer) appendTrace(path string, start time.Time, duration time.Duration, status int) {
	t.Lock()
	defer t.Unlock()

	t.traceList.list = append(t.traceList.list, trace{
		Path:     path,
		Start:    start,
		Duration: duration.Seconds(),
		Status:   status,
	})

	if len(t.traceList.list) >= maxTraces || time.Since(t.traceList.createdAt) > flushTime {
		var oldList = t.traceList
		t.traceList = newTraceList()
		writeTraces(oldList.list)
	}
}

func (t *tracer) shutdown(wg *sync.WaitGroup) {
	t.Lock()
	writeTraces(t.traceList.list)
	t.Unlock()
	wg.Done()
}

type registry struct {
	sync.Mutex
	list []*tracer
}

func (r *registry) new(h http.Handler) *tracer {
	var t = &tracer{handler: h, traceList: newTraceList()}
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
