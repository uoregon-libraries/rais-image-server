package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"rais/src/iiif"
	"reflect"
	"sort"
	"strings"

	"github.com/uoregon-libraries/gopkg/logger"
)

var wrapHandlerPlugins []func(string, http.Handler) (http.Handler, error)
var teardownPlugins []func()
var purgeCachePlugins []func()
var expireCachedImagePlugins []func(iiif.ID)

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
			l.Warnf("Skipping invalid plugin pattern %q: %s", pattern, err)
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
		var err = loadPlugin(file, l)
		if err != nil {
			l.Errorf("Unable to load %q: %s", file, err)
		}
	}
}

type pluginWrapper struct {
	*plugin.Plugin
	path      string
	functions []string
	errors    []string
}

func newPluginWrapper(path string) (*pluginWrapper, error) {
	var p, err = plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot load plugin %q: %s", path, err)
	}
	return &pluginWrapper{Plugin: p, path: path}, nil
}

// loadPluginFn loads the symbol by the given name and attempts to set it to
// the given object via reflection.  If the two aren't the same type, an error
// is added to the pluginWrapper's error list.
func (pw *pluginWrapper) loadPluginFn(name string, obj any) {
	var sym, err = pw.Lookup(name)
	if err != nil {
		return
	}

	var objElem = reflect.ValueOf(obj).Elem()
	var objType = objElem.Type()
	var symV = reflect.ValueOf(sym)

	if !symV.Type().AssignableTo(objType) {
		pw.errors = append(pw.errors, fmt.Sprintf("invalid signature for %s (expecting %s)", name, objType))
		return
	}

	objElem.Set(symV)
	pw.functions = append(pw.functions, name)
}

// loadPlugin attempts to read the given plugin file and extract known symbols.
// If a plugin exposes Initialize or SetLogger, they're called here once we're
// sure the plugin is valid.  All other functions are indexed globally for use
// in the RAIS image serving handler.
func loadPlugin(fullpath string, l *logger.Logger) error {
	var pw, err = newPluginWrapper(fullpath)
	if err != nil {
		return err
	}

	// Set up dummy / no-op functions so we can call these without risk
	var log = func(*logger.Logger) {}
	var initialize = func() {}

	// Simply initialize those functions we only want indexed if they exist
	var teardown func()
	var wrapHandler func(string, http.Handler) (http.Handler, error)
	var prgCache func()
	var expCachedImg func(iiif.ID)

	pw.loadPluginFn("SetLogger", &log)
	pw.loadPluginFn("Initialize", &initialize)
	pw.loadPluginFn("Teardown", &teardown)
	pw.loadPluginFn("WrapHandler", &wrapHandler)
	pw.loadPluginFn("PurgeCaches", &prgCache)
	pw.loadPluginFn("ExpireCachedImage", &expCachedImg)

	if len(pw.errors) != 0 {
		return errors.New(strings.Join(pw.errors, ", "))
	}
	if len(pw.functions) == 0 {
		return fmt.Errorf("no known functions exposed")
	}

	// We need to call SetLogger and Initialize immediately, as they're never
	// called a second time and they tell us if the plugin is going to be used
	log(l)
	initialize()

	// After initialization, we check if the plugin explicitly set itself to Disabled
	var sym plugin.Symbol
	sym, err = pw.Lookup("Disabled")
	if err == nil {
		var disabled, ok = sym.(*bool)
		if !ok {
			return fmt.Errorf("non-boolean Disabled value exposed")
		}
		if *disabled {
			l.Infof("%q is disabled", fullpath)
			return nil
		}
		l.Debugf("%q is explicitly enabled", fullpath)
	}

	// Index remaining functions
	if teardown != nil {
		teardownPlugins = append(teardownPlugins, teardown)
	}
	if wrapHandler != nil {
		wrapHandlerPlugins = append(wrapHandlerPlugins, wrapHandler)
	}
	if prgCache != nil {
		purgeCachePlugins = append(purgeCachePlugins, prgCache)
	}
	if expCachedImg != nil {
		expireCachedImagePlugins = append(expireCachedImagePlugins, expCachedImg)
	}

	// Add info to stats
	stats.Plugins = append(stats.Plugins, plugStats{
		Path:      fullpath,
		Functions: pw.functions,
	})

	return nil
}
