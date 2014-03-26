#include <stdio.h>
#include <openjpeg-2.0/openjpeg.h>
#include "handlers.h"
#include "_cgo_export.h"

static void info_callback(const char *msg, void *client_data) {
	(void)client_data;
	fprintf(stdout, "[INFO] %s", msg);
}

static void warning_callback(const char *msg, void *client_data) {
	(void)client_data;
	fprintf(stdout, "[WARNING] %s", msg);
}

static void error_callback(const char *msg, void *client_data) {
	(void)client_data;
	fprintf(stdout, "[ERROR] %s", msg);
}

void set_handlers(opj_codec_t * p_codec) {
	opj_set_info_handler(p_codec, info_callback, 00);
	opj_set_warning_handler(p_codec, warning_callback, 00);
	opj_set_error_handler(p_codec, error_callback, 00);
}
