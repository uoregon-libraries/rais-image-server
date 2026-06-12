#include <assert.h>
#include "magick.h"

void SetImageInfoFilename(ImageInfo *image_info, char *filename) {
  (void) CopyMagickString(image_info->filename,filename,MagickPathExtent);
}

int HasError(ExceptionInfo *exception) {
  assert(exception != (ExceptionInfo *) NULL);
  assert(exception->signature == MagickCoreSignature);

  if (exception->severity >= ErrorException)
    return 1;

  return 0;
}

void ExportRGBA(Image *image, size_t w, size_t h, void *pixels, ExceptionInfo *e) {
	ExportImagePixels(image, 0, 0, w, h, "RGBA", CharPixel, pixels, e);
}

RectangleInfo MakeRectangle(int x, int y, int w, int h) {
  RectangleInfo ri;
  ri.x = x;
  ri.y = y;
  ri.width = w;
  ri.height = h;

  return ri;
}

Image *Resize(Image *image, size_t w, size_t h, ExceptionInfo *e) {
  return AdaptiveResizeImage(image, w, h, e);
}
