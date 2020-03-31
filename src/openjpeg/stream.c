#include <stdio.h>
#include <openjpeg.h>
#include "stream.h"
#include "_cgo_export.h"

OPJ_SIZE_T stream_read(void * p_buffer, OPJ_SIZE_T p_nb_bytes, void *stream_id) {
  return opjStreamRead(p_buffer, p_nb_bytes, (OPJ_UINT64)stream_id);
}

OPJ_OFF_T stream_skip(OPJ_OFF_T p_nb_bytes, void *stream_id) {
  return opjStreamSkip(p_nb_bytes, (OPJ_UINT64)stream_id);
}

OPJ_BOOL stream_seek(OPJ_OFF_T p_nb_bytes, void *stream_id) {
  return opjStreamSeek(p_nb_bytes, (OPJ_UINT64)stream_id);
}

void free_stream(void *stream_id) {
  freeStream((OPJ_UINT64)stream_id);
}

opj_stream_t* new_stream(OPJ_UINT64 buffer_size, OPJ_UINT64 stream_id, OPJ_UINT64 data_size) {
    opj_stream_t* l_stream = 00;

    l_stream = opj_stream_create(buffer_size, 1);
    if (! l_stream) {
        return NULL;
    }

    opj_stream_set_user_data(l_stream, (void*)stream_id, free_stream);
    opj_stream_set_user_data_length(l_stream, data_size);
    opj_stream_set_read_function(l_stream, (opj_stream_read_fn) stream_read);
    opj_stream_set_skip_function(l_stream, (opj_stream_skip_fn) stream_skip);
    opj_stream_set_seek_function(l_stream, (opj_stream_seek_fn) stream_seek);

    return l_stream;
}
