package main

import(
	"fmt"
	"image"
	"github.com/eikeon/brikker/openjpeg"
)

// Image tile-pulling values
var r image.Rectangle = image.Rect(100, 100, 358, 358)
var width int = 127
var height int = 127

// Wrapper to doVerify which sets up channel stuff so we can let the main app
// know when processing of this file is done
func verifyJP2(path string) {
	message := doVerify(path)
	if message != "" {
		jp2Messages <- message
	}
}

func doVerify(path string) string {
	message := fmt.Sprintf("Opening %#v: ", path)

	err, _ := openjpeg.NewImageTile(path, r, width, height)

	if (err != nil) {
		message += fmt.Sprintf("Error creating image tile: %s", err)
		return message
	}

	message += "Success!"
	return message
}

func createWorker() {
	for {
		verifyJP2(<-jp2Files)
	}
}
