package main

import(
	"fmt"
	"os"
	"github.com/eikeon/brikker/openjpeg"
)

// Buffer for holding all the output information as jp2s are examined
var jp2Messages chan string

// Channel for holding the files our workers need to process
var jp2Files chan string

func main() {
	// This seems like a solid value based on various bits of testing, but it may
	// make sense to do something smarter here
	maxWorkers := 8

	// Set up channels with big enough buffers that one goroutine isn't holding
	// up another
	jp2Messages = make(chan string, maxWorkers)
	jp2Files = make(chan string, maxWorkers)

	// Read JP2 file list and stuff it into the channel
	if (len(os.Args) != 2) {
		fmt.Println("You must provide a path to a file listing JP2s to verify")
		os.Exit(1)
	}
	fileCount := readJP2FileList(os.Args[1])

	// Show anything that's "info" or more serious
	openjpeg.LogLevel = 4

	fmt.Println("BEGIN")
	fmt.Println("---")

	for i := 0; i < maxWorkers; i++ {
		go createWorker()
	}

	// Read messages until we have accounted for all JP2s
	for i := 0; i < fileCount; i++ {
		message := <-jp2Messages
		fmt.Println(message)
	}

	fmt.Println("---")
	fmt.Println("COMPLETE")
}
