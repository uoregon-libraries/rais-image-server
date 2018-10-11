package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"rais/src/iiif"
	"sort"

	"github.com/uoregon-libraries/gopkg/logger"
)

type plugGeneric func()
type plugIDToPath func(iiif.ID) (string, error)
type plugLogger func(*logger.Logger)
type plugWrapHandler func(string, http.Handler) (http.Handler, error)

var idToPathPlugins []plugIDToPath
var wrapHandlerPlugins []plugWrapHandler
var teardownPlugins []plugGeneric

// LoadPlugins searches for any plugins in the binary's directory + /plugins
func LoadPlugins(l *logger.Logger) {
	var dir = filepath.Dir(os.Args[0])
	var plugdir = filepath.Join(dir, "plugins")
	l.Debugf("Looking for plugins in %q", plugdir)

	var _, err = os.Stat(plugdir)
	if os.IsNotExist(err) {
		// If there's no plugin dir, we just keep chugging along normally
		l.Debugf("Plugin directory not found; skipping plugin loading")
		return
	}
	if err != nil {
		l.Fatalf("Unable to stat %q: %s", plugdir, err)
	}

	var infos []os.FileInfo
	infos, err = ioutil.ReadDir(plugdir)
	if err != nil {
		l.Fatalf("Unable to read plugin directory %q: %s", plugdir, err)
	}

	sort.Slice(infos, func(i, j int) bool { return infos[i].Name() < infos[j].Name() })
	for _, info := range infos {
		var fullpath = filepath.Join(plugdir, info.Name())
		if info.IsDir() {
			l.Warnf("Skipping unknown subdirectory %q (plugin subdirectories are not supported)", fullpath)
		}

		if filepath.Ext(fullpath) != ".so" {
			l.Warnf("Skipping unknown file %q (plugins must be compiled .so files)", fullpath)
		}

		l.Infof("Loading plugin %q", fullpath)
		loadPlugin(fullpath, l)
	}
}

// loadPlugin attempts to read the given plugin file and extract known symbols.
// If a plugin exposes Initialize or SetLogger, they're called here once we're
// sure the plugin is valid.  IDToPath functions are indexed globally for use
// in the RAIS image serving handler.
//
// This function is unnecessarily complicated and needs refactoring.  Other
// than the concrete type, the "index a function" blocks are all basically the
// same.
func loadPlugin(fullpath string, l *logger.Logger) {
	var p, err = plugin.Open(fullpath)
	if err != nil {
		l.Errorf("Error loading plugin %q: %s", fullpath, err)
		return
	}

	var sym plugin.Symbol
	var fnCount int

	var log plugLogger
	sym, err = p.Lookup("SetLogger")
	if err == nil {
		var f, ok = sym.(func(*logger.Logger))
		if !ok {
			l.Errorf("%q.SetLogger is invalid", fullpath)
			return
		}
		l.Debugf("Found %q.SetLogger", fullpath)
		log = plugLogger(f)
		fnCount++
	}

	var idToPath plugIDToPath
	sym, err = p.Lookup("IDToPath")
	if err == nil {
		var f, ok = sym.(func(iiif.ID) (string, error))
		if !ok {
			l.Errorf("%q.IDToPath is invalid", fullpath)
			return
		}
		l.Debugf("Found %q.IDToPath", fullpath)
		idToPath = plugIDToPath(f)
		fnCount++
	}

	var init plugGeneric
	sym, err = p.Lookup("Initialize")
	if err == nil {
		var f, ok = sym.(func())
		if !ok {
			l.Errorf("%q.Initialize is invalid", fullpath)
			return
		}
		l.Debugf("Found %q.Initialize", fullpath)
		init = plugGeneric(f)
		fnCount++
	}

	var teardown plugGeneric
	sym, err = p.Lookup("Teardown")
	if err == nil {
		var f, ok = sym.(func())
		if !ok {
			l.Errorf("%q.Teardown is invalid", fullpath)
			return
		}

		l.Debugf("Found %q.Teardown", fullpath)
		teardown = plugGeneric(f)
		fnCount++
	}

	var wrapHandler plugWrapHandler
	sym, err = p.Lookup("WrapHandler")
	if err == nil {
		var f, ok = sym.(func(string, http.Handler) (http.Handler, error))
		if !ok {
			l.Errorf("%q.WrapHandler is invalid", fullpath)
			return
		}

		l.Debugf("Found %q.WrapHandler", fullpath)
		wrapHandler = plugWrapHandler(f)
		fnCount++
	}

	if fnCount == 0 {
		l.Warnf("%q doesn't expose any known functions", fullpath)
		return
	}

	// We can call SetLogger and Initialize immediately, as they're never called a second time
	if log != nil {
		log(l)
	}
	if init != nil {
		init()
	}

	// Index other available functions
	if idToPath != nil {
		idToPathPlugins = append(idToPathPlugins, idToPath)
	}
	if teardown != nil {
		teardownPlugins = append(teardownPlugins, teardown)
	}
	if wrapHandler != nil {
		wrapHandlerPlugins = append(wrapHandlerPlugins, wrapHandler)
	}
}
