package main

import (
	"fmt"
	"github.com/uoregon-libraries/newspaper-jp2-viewer/openjpeg"
	"image"
	"os"
)

// Image tile-pulling values
var r image.Rectangle = image.Rect(100, 100, 358, 358)
var width int = 127
var height int = 127

// Verifies that file exists
func checkFile(path string) error {
	_, err := os.Stat(path)
	return err
}

// Wrapper to doVerify which tests file prior to doVerify call, and pushes any
// errors onto the jp2Messages channel
func verifyJP2(path string) {
	if err := checkFile(path); err != nil {
		jp2Messages <- fmt.Sprintf("ERROR reading JP2 file: %s", err)
		return
	}

	jp2Messages <- doVerify(path)
}

// Verifies that we can read and serve tiles for the given JP2.  This
// effectively determines if the installed openjpeg libs will work.
func doVerify(path string) string {
	jp2, err := openjpeg.NewJP2Image(path)
	if err != nil {
		return fmt.Sprintf("ERROR reading JP2 image %#v: %s", path, err)
	}

	jp2.SetCrop(r)
	jp2.SetResize(width, height)

	_, err = jp2.DecodeImage()

	if err != nil {
		return fmt.Sprintf("ERROR creating image tile from %#v: %s", path, err)
	}

	return fmt.Sprintf("SUCCESS: %#v", path)
}

func createWorker() {
	for {
		verifyJP2(<-jp2Files)
	}
}
