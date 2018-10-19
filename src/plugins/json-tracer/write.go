package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"strconv"
	"time"

	"github.com/uoregon-libraries/gopkg/fileutil"
)

func writeEvents(list []event) {
	if len(list) == 0 {
		return
	}

	// We build the filename from a nanosecond-level timestamp and a random
	// integer to make collisions effectively impossible
	var now = time.Now()
	var ts = now.Format("2006-01-02_15-04-05_") + fmt.Sprintf("%09d", now.Nanosecond())
	var rnd = rand.Intn(math.MaxInt32)
	var rndS = strconv.Itoa(int(1e9 + rnd%1e9))[1:]
	var fullpath = filepath.Join(jsonOutDir, ts+rndS)

	var f = fileutil.NewSafeFile(fullpath)
	var bytes, err = json.Marshal(list)
	if err != nil {
		l.Criticalf("Unrecoverable instrumentation failure: unable to marshal event data: %s", err)
		f.Cancel()
		return
	}

	tryWrite(f, bytes)
}

func tryWrite(f fileutil.WriteCancelCloser, bytes []byte) {
	var try uint
	for try = 0; try < 8; try++ {
		var _, err = f.Write(bytes)
		if err == nil {
			tryClose(f)
			return
		}

		var backoff = 1 << try
		l.Warnf("Unable to write instrumentation data to file; trying again in %d seconds.  Error: %s", backoff, err)
		time.Sleep(time.Duration(backoff) * time.Second)
	}

	l.Criticalf("Unrecoverable instrumentation failure: unable to write event data after many attempts")
	f.Cancel()
}

func tryClose(f fileutil.WriteCancelCloser) {
	var try uint
	for try = 0; try < 8; try++ {
		var err = f.Close()
		if err == nil {
			return
		}

		var backoff = 1 << try
		l.Warnf("Unable to close instrumentation file; trying again in %d seconds.  Error: %s", backoff, err)
		time.Sleep(time.Duration(backoff) * time.Second)
	}

	l.Criticalf("Unrecoverable instrumentation failure: unable to close event data file after many attempts")
	f.Cancel()
}
