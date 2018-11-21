// purge.go does a really horrible job of purging files that haven't been read
// by RAIS in a certain timeframe.  If we like this approach, it needs to be
// extracted into a central package usable from other plugins, and maybe even
// something completely external so it's useful from other projects.

package main

import (
	"time"
)

// purgeLoop checks if cached files need to be purged every few seconds
func purgeLoop() {
	for {
		checkPurge()
		time.Sleep(time.Second * 5)
	}
}

func checkPurge() {
	for _, a := range assets {
		if time.Now().Add(-cacheLifetime).Before(a.lastAccess) {
			go doPurge(a)
		}
	}
}

func doPurge(a *asset) {
	a.purge()
	assetMutex.Lock()
	delete(assets, a.id)
	assetMutex.Unlock()
}
