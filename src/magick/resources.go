package magick

/*
#cgo pkg-config: MagickCore
#include <magick/MagickCore.h>
*/
import "C"

func finalizer(i *Image) {
	i.CleanupResources()
}

func (i *Image) CleanupImage() {
	if i.image != nil {
		if i.image.next != nil {
			C.DestroyImageList(i.image)
		} else {
			C.DestroyImage(i.image)
		}
		i.image = nil
	}
}

func (i *Image) CleanupImageInfo() {
	if i.imageInfo != nil {
		C.DestroyImageInfo(i.imageInfo)
		i.imageInfo = nil
	}
}

func (i *Image) CleanupResources() {
	i.CleanupImage()
	i.CleanupImageInfo()
}
