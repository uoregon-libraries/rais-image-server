package main

import(
	"fmt"
	"image"
	"os"
	"github.com/eikeon/brikker/openjpeg"
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

// Verifies a JP2 file can have a tile read by Brikker and openjpeg
func doVerify(path string) string {
	err, _ := openjpeg.NewImageTile(path, r, width, height)

	if (err != nil) {
		return fmt.Sprintf("ERROR creating image tile from %#v: %s", path, err)
	}

	return fmt.Sprintf("SUCCESS: %#v", path)
}

func createWorker() {
	for {
		verifyJP2(<-jp2Files)
	}
}
