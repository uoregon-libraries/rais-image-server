package magick

/*
#cgo pkg-config: MagickCore
#include <magick/MagickCore.h>
*/
import "C"

func finalizer(i *Image) {
	i.CleanupResources()
}

func (i *Image) cleanupImage() {
	if i.image != nil {
		if i.image.next != nil {
			C.DestroyImageList(i.image)
		} else {
			C.DestroyImage(i.image)
		}
		i.image = nil
	}
}

func (i *Image) cleanupImageInfo() {
	if i.imageInfo != nil {
		C.DestroyImageInfo(i.imageInfo)
		i.imageInfo = nil
	}
}

// CleanupResources frees the C data allocated by ImageMagick
func (i *Image) CleanupResources() {
	i.cleanupImage()
	i.cleanupImageInfo()
}
