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

  LockSemaphoreInfo(exception->semaphore);
  ResetLinkedListIterator((LinkedListInfo *) exception->exceptions);
  p=(const ExceptionInfo *) GetNextValueInLinkedList((LinkedListInfo *)
    exception->exceptions);
  while (p != (const ExceptionInfo *) NULL) {
    if ((p->severity >= WarningException) && (p->severity < ErrorException))
      result = 1;
    if ((p->severity >= ErrorException) && (p->severity < FatalErrorException))
      result = 1;
    if (p->severity >= FatalErrorException)
      result = 1;
    p=(const ExceptionInfo *) GetNextValueInLinkedList((LinkedListInfo *)
      exception->exceptions);
  }
  UnlockSemaphoreInfo(exception->semaphore);
	return result;
}

Image *ReadImageFromBlob(ImageInfo *image_info, void *blob, size_t length) {
  Image *image;
  ExceptionInfo *exception;
  exception = AcquireExceptionInfo();
  image_info->blob = blob;
  image_info->length = length;
  image = ReadImage(image_info, exception);
  CatchException(exception);
  DestroyExceptionInfo(exception);
  return image;
}
