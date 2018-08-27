package main

import (
	"fmt"
	"os"
	"rais/src/jp2info"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Raw bool `short:"r" long:"raw" description:"show raw JP2 info structure"`
}

func main() {
	var args []string
	var err error

	var parser = flags.NewParser(&opts, flags.Default)
	parser.Usage = "filename [filename...] [OPTIONS]"
	args, err = parser.Parse()

	if err != nil || len(args) < 1 {
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	var arg string
	var s = new(jp2info.Scanner)
	for _, arg = range args {
		fmt.Printf("%s: ", arg)
		printScanResults(s.Scan(arg))
	}
}

func printScanResults(i *jp2info.Info, err error) {
	if err != nil {
		// Invalid file or some variation of the spec that isn't supported
		fmt.Printf("Error: %s\n", err)
		return
	}

	if opts.Raw {
		fmt.Printf("%#v\n", i)
	} else {
		printInfo(i)
	}
}

func printInfo(i *jp2info.Info) {
	fmt.Printf("dim:%dx%d tiles:%dx%d levels:%d %s\n",
		i.Width, i.Height, i.TileWidth(), i.TileHeight(), i.Levels, i.ColorSpace.String())
}
