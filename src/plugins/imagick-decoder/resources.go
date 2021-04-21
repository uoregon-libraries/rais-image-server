package main

/*
#cgo pkg-config: MagickCore
#include <magick/MagickCore.h>
*/
import "C"

func cleanupImage(i *C.Image) {
	if i != nil {
		if i.next != nil {
			C.DestroyImageList(i)
		} else {
			C.DestroyImage(i)
		}
	}
}

func cleanupImageInfo(i *C.ImageInfo) {
	if i != nil {
		C.DestroyImageInfo(i)
	}
}
