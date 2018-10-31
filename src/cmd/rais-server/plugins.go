package main

import (
	"fmt"
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

// pluginsFor returns a list of all plugin files which matched the given
// pattern.  Files are sorted by name.
func pluginsFor(pattern string) ([]string, error) {
	if !filepath.IsAbs(pattern) {
		var dir = filepath.Join(filepath.Dir(os.Args[0]), "plugins")
		pattern = filepath.Join(dir, pattern)
	}

	var files, err = filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin file pattern %q", pattern)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("plugin pattern %q doesn't match any files", pattern)
	}

	sort.Strings(files)
	return files, nil
}

// LoadPlugins searches for any plugins matching the pattern given.  If the
// pattern is not an absolute URL, it is treated as a pattern under the
// binary's dir/plugins.
func LoadPlugins(l *logger.Logger, patterns []string) {
	var plugFiles []string
	var seen = make(map[string]bool)
	for _, pattern := range patterns {
		var matches, err = pluginsFor(pattern)
		if err != nil {
			l.Fatalf("Cannot process pattern %q: %s", pattern, err)
		}

		// We do a sanity check before actually processing any plugins
		for _, file := range matches {
			if filepath.Ext(file) != ".so" {
				l.Fatalf("Cannot load unknown file %q (plugins must be compiled .so files)", file)
			}
			if seen[file] {
				l.Fatalf("Cannot load the same plugin twice (%q)", file)
			}
			seen[file] = true
		}

		plugFiles = append(plugFiles, matches...)
	}

	for _, file := range plugFiles {
		l.Infof("Loading plugin %q", file)
		loadPlugin(file, l)
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

	// After initialization, we check if the plugin explicitly set itself to Disabled
	sym, err = p.Lookup("Disabled")
	if err == nil {
		var disabled, ok = sym.(*bool)
		if !ok {
			l.Errorf("%q.Disabled is not a boolean", fullpath)
			return
		}
		if *disabled {
			l.Infof("%q is disabled", fullpath)
			return
		}
		l.Debugf("%q is explicitly enabled", fullpath)
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
