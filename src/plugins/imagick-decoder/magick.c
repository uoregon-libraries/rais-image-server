#include <assert.h>
#include "magick.h"

void SetImageInfoFilename(ImageInfo *image_info, char *filename) {
  (void) CopyMagickString(image_info->filename,filename,MaxTextExtent);
}

int HasError(ExceptionInfo *exception) {
  register const ExceptionInfo *p;
  int result = 0;

  assert(exception != (ExceptionInfo *) NULL);
  assert(exception->signature == MagickSignature);
  if (exception->exceptions  == (void *) NULL)
    return 0;
  if (exception->semaphore == (void *) NULL)
    return 0;

  LockSemaphoreInfo(exception->semaphore);
  ResetLinkedListIterator((LinkedListInfo *) exception->exceptions);
  p=(const ExceptionInfo *) GetNextValueInLinkedList((LinkedListInfo *)
    exception->exceptions);
  while (p != (const ExceptionInfo *) NULL) {
    if (p->severity >= ErrorException)
      result = 1;
    p=(const ExceptionInfo *) GetNextValueInLinkedList((LinkedListInfo *)
      exception->exceptions);
  }
  UnlockSemaphoreInfo(exception->semaphore);
	return result;
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
