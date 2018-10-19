package main

import (
	"net/http"
	"rais/src/iiif"
	"strings"
	"sync"
	"time"
)

type event struct {
	Path     string
	Type     string
	Start    time.Time
	Duration float64
	Status   int
}

type eventList struct {
	createdAt time.Time
	list      []event
}

func newEventList() *eventList {
	return &eventList{list: make([]event, 0, 256)}
}

type tracer struct {
	sync.Mutex
	done    bool
	handler http.Handler
	events  *eventList
}

// ServeHTTP implements http.Handler.  We call the underlying handler and store
// timing data locally.
func (t *tracer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var sr = statusRecorder{w, 200}
	var path = req.URL.RawPath
	if path == "" {
		path = req.URL.Path
	}

	var start = time.Now()
	t.handler.ServeHTTP(&sr, req)
	var finish = time.Now()

	// To avoid blocking when the events are being processed, we send the event
	// to the tracer's list asynchronously
	go t.appendEvent(path, start, finish, sr.status)
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

func (t *tracer) appendEvent(path string, start, finish time.Time, status int) {
	t.Lock()
	defer t.Unlock()

	if len(t.events.list) == 0 {
		t.events.createdAt = time.Now()
	}
	t.events.list = append(t.events.list, event{
		Path:     path,
		Type:     getReqType(path),
		Start:    start,
		Duration: finish.Sub(start).Seconds(),
		Status:   status,
	})
}

// loop checks regularly for the last flush having been long enough ago to
// flush to disk again.  This must run in a background goroutine.
func (t *tracer) loop() {
	for {
		var pending []event
		t.Lock()
		var tlen = len(t.events.list)
		if tlen > 0 && time.Since(t.events.createdAt) > flushTime {
			pending = t.events.list
			t.events = newEventList()
		}
		var done = t.done
		t.Unlock()

		if len(pending) > 0 {
			writeEvents(pending)
		}

		if done {
			return
		}

		time.Sleep(5 * time.Second)
	}
}

func (t *tracer) shutdown(wg *sync.WaitGroup) {
	t.Lock()
	writeEvents(t.events.list)
	t.done = true
	t.Unlock()
	wg.Done()
}

type registry struct {
	list []*tracer
}

func (r *registry) new(h http.Handler) *tracer {
	var t = &tracer{handler: h, events: newEventList()}
	go t.loop()
	r.list = append(r.list, t)
	return t
}

func (r *registry) shutdown() {
	var wg sync.WaitGroup
	for _, t := range r.list {
		wg.Add(1)
		go t.shutdown(&wg)
	}

	wg.Wait()
}
