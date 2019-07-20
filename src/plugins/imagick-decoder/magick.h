#include <assert.h>
#include <magick/MagickCore.h>

extern void SetImageInfoFilename(ImageInfo *image_info, char *filename);
extern int HasError(ExceptionInfo *exception);
extern void ExportRGBA(Image *image, size_t w, size_t h, void *pixels, ExceptionInfo *e);
extern RectangleInfo MakeRectangle(int x, int y, int w, int h);
extern Image *Resize(Image *image, size_t w, size_t h, ExceptionInfo *e);
