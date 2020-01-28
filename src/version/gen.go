// +build ignore

package main

import (
	"os"
	"os/exec"
	"rais/src/version"
	"strings"
)

func main() {
	var err error

	var cmd = exec.Command("git", "describe")
	var out []byte
	out, err = cmd.CombinedOutput()

	// This can fail when there's no git repository, so instead of crashing, we
	// just have a build tag of "indev"
	var build string
	if err == nil {
		build = string(out)
	} else {
		build = version.Version + "-indev"
	}

	build = strings.TrimSpace(build)

	var f *os.File
	f, err = os.Create("build.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteString(`// build.go is a generated file and should not be modified by hand

package version

const Build = "` + build + `"
`)
	if err != nil {
		panic(err)
	}
}
