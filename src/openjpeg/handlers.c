#include <stdio.h>
#include <openjpeg.h>
#include "handlers.h"
#include "_cgo_export.h"

static void warning_callback(const char *msg, void *client_data) {
	GoLogWarning((char *)msg);
}

static void error_callback(const char *msg, void *client_data) {
	GoLogError((char *)msg);
}

void set_handlers(opj_codec_t* p_codec) {
	opj_set_warning_handler(p_codec, warning_callback, 00);
	opj_set_error_handler(p_codec, error_callback, 00);
}
