#include <assert.h>
#include <magick/MagickCore.h>

extern void SetImageInfoFilename(ImageInfo *image_info, char *filename);
extern int HasError(ExceptionInfo *exception);
extern Image *ReadImageFromBlob(ImageInfo *image_info, void *blob, size_t length);
