// purge.go does a really horrible job of purging files that haven't been read
// by RAIS in a certain timeframe.  If we like this approach, it needs to be
// extracted into a central package usable from other plugins, and maybe even
// something completely external so it's useful from other projects.

package main

import (
	"os"
	"sync"
	"time"
)

var purgeM sync.Mutex
var purges = make(map[string]time.Time)

func setPurgeTime(path string) {
	purgeM.Lock()
	purges[path] = time.Now().Add(cacheLifetime)
	purgeM.Unlock()
}

// purgeLoop checks if cached files need to be purged every few seconds
func purgeLoop() {
	for {
		checkPurge()
		time.Sleep(time.Second * 5)
	}
}

func checkPurge() {
	purgeM.Lock()
	defer purgeM.Unlock()

	var removed []string
	for path, when := range purges {
		if time.Now().After(when) {
			var err = os.Remove(path)
			if err != nil {
				l.Errorf("s3-images plugin: Unable to purge cached file at %q: %s", path, err)
				continue
			}
			l.Infof("s3-images plugin: Purged %q", path)
			removed = append(removed, path)
		}
	}

	for _, path := range removed {
		delete(purges, path)
	}
}
