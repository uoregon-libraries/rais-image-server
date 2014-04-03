package main

import(
	"os"
	"fmt"
	"io/ioutil"
	"strings"
)

// Populates the jp2Files channel
func pushJP2Filenames(filenames []string) {
	for _, file := range filenames {
		jp2Files <- file
	}
}

// Reads user-specified JP2 file list and pushes files into the jp2 file
// channel for workers to process
func readJP2FileList(jp2ListFile string) (count int) {
	// Read data from file
	fmt.Println("Attempting to read from", jp2ListFile)
	content, err := ioutil.ReadFile(jp2ListFile)
	if err != nil {
		fmt.Println("Unable to read JP2 list file! (", err, ")")
		os.Exit(1)
	}

	fileList := strings.Split(string(content), "\n")
	validFiles := make([]string, 0)

	for _, file := range fileList {
		if file != "" {
			validFiles = append(validFiles, file)
		}
	}

	go pushJP2Filenames(validFiles)
	return len(validFiles)
}
