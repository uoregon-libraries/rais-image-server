package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"rais/src/iiif"

	"github.com/uoregon-libraries/gopkg/logger"
)

type plugIDToPath func(iiif.ID) (string, error)

var idToPathPlugins []plugIDToPath

// LoadPlugins searches for any plugins in the current working directory /plugins
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

func loadPlugin(fullpath string, l *logger.Logger) {
	var p, err = plugin.Open(fullpath)
	if err != nil {
		l.Errorf("Error loading plugin %q: %s", fullpath, err)
		return
	}

	var sym plugin.Symbol

	sym, err = p.Lookup("SetLogger")
	if err == nil {
		var setLogger, ok = sym.(func(*logger.Logger))
		if ok {
			l.Debugf("Attaching central logger to plugin")
			setLogger(l)
		}
	}

	sym, err = p.Lookup("IDToPath")
	if err == nil {
		var f, ok = sym.(func(iiif.ID) (string, error))
		if !ok {
			l.Errorf("Plugin %q exposes an invalid IDToPath function", fullpath)
			return
		}

		l.Debugf("Adding IDToPath from %q", fullpath)
		idToPathPlugins = append(idToPathPlugins, plugIDToPath(f))
	}
}
