package main

import (
	"encoding/json"
	"os"
	"time"
)

// ready returns true if the tracer is ready to flush events to disk
func (t *tracer) ready() bool {
	t.Lock()
	defer t.Unlock()
	return time.Now().After(t.nextFlushTime) && len(t.events) > 0
}

func (t *tracer) flush() {
	t.Lock()
	defer t.Unlock()

	t.nextFlushTime = time.Now().Add(flushTime)

	// This is only necessary for forced flushing, such as on shutdown
	if len(t.events) == 0 {
		return
	}

	// Generate the JSON output first so we can report truly fatal errors before bothering with file IO
	var towrite []byte
	for _, ev := range t.events {
		var bytes, err = json.Marshal(ev)
		if err != nil {
			l.Errorf("json-tracer plugin: skipping 1 event: unable to marshal event (%#v) data: %s", ev, err)
			continue
		}
		towrite = append(towrite, bytes...)
		towrite = append(towrite, '\n')
	}

	var f, err = os.OpenFile(jsonOut, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		l.Errorf("json-tracer plugin: unable to open %q for appending: %s", jsonOut, err)
		t.addWriteFailure()
		return
	}
	defer f.Close()

	_, err = f.Write(towrite)
	if err != nil {
		l.Errorf("json-tracer plugin: unable to write events: %s", err)
		t.addWriteFailure()
		return
	}

	err = f.Close()
	if err != nil {
		l.Errorf("json-tracer plugin: unable to close %q: %s", jsonOut, err)
		t.addWriteFailure()
		return
	}

	if t.writeFailures > 0 {
		l.Infof("json-tracer plugin: successfully wrote held events")
	}
	t.events = makeEvents()
	t.writeFailures = 0
}

func (t *tracer) addWriteFailure() {
	// Let's max out at 8 failures to avoid waiting so long that the next flush never happens
	if t.writeFailures < 8 {
		t.writeFailures++
	}

	// We'll forcibly change the next flush attempt so we can recover quickly if
	// the problem is short-lived, but try less and less often the more failures
	// we've logged
	var sec = 1 << uint(t.writeFailures)
	t.nextFlushTime = time.Now().Add(time.Second * time.Duration(sec))
	l.Warnf("json-tracer plugin: next flush() in %d second(s)", sec)

	// If we've been failing for a while and we have over 10k stored events, we start dropping some....
	var elen = len(t.events)
	if t.writeFailures > 5 && elen > 10000 {
		t.events = t.events[elen-10000:]
		l.Criticalf("json-tracer plugin: continued write failures has resulted in dropping %d events", elen-len(t.events))
	}
}
