package openjpeg

// #cgo pkg-config: libopenjp2
// #include <openjpeg.h>
import "C"
import (
	"io"
	"reflect"
	"sync"
	"unsafe"
)

// These vars suck, but we have to have some way to find our objects when the
// openjpeg C callbacks tell us to read/seek/etc.  Trying to persist a pointer
// between Go and C is definitely more dangerous than this hack....
var nextStreamID uint64
var images = make(map[uint64]*JP2Image)
var imageMutex sync.RWMutex

// These are stupid, but we need to return what openjpeg considers failure
// numbers, and Go doesn't allow a direct translation of negative values to an
// unsigned type
var opjZero64 C.OPJ_UINT64 = 0
var opjMinusOne64 = opjZero64 - 1
var opjZeroSizeT C.OPJ_SIZE_T = 0
var opjMinusOneSizeT = opjZeroSizeT - 1

// storeImage stores the next sequence id on the JP2 image and indexes it in
// our image lookup map so opj streaming functions can find it
func storeImage(i *JP2Image) {
	imageMutex.Lock()
	nextStreamID++
	i.id = nextStreamID
	images[i.id] = i
	imageMutex.Unlock()
}

func lookupImage(id uint64) (*JP2Image, bool) {
	imageMutex.Lock()
	var i, ok = images[id]
	imageMutex.Unlock()

	return i, ok
}

//export freeStream
func freeStream(id uint64) {
	imageMutex.Lock()
	delete(images, id)
	imageMutex.Unlock()
}

//export opjStreamRead
func opjStreamRead(writeBuffer unsafe.Pointer, numBytes C.OPJ_SIZE_T, id uint64) C.OPJ_SIZE_T {
	var i, ok = lookupImage(id)
	if !ok {
		Logger.Errorf("Unable to find stream %d", id)
		return opjMinusOneSizeT
	}

	var data []byte
	var dataSlice = (*reflect.SliceHeader)(unsafe.Pointer(&data))
	dataSlice.Cap = int(numBytes)
	dataSlice.Len = int(numBytes)
	dataSlice.Data = uintptr(unsafe.Pointer(writeBuffer))

	var n, err = i.streamer.Read(data)

	// Dumb hack - gocloud (maybe others?) returns EOF differently for local file
	// read vs. an S3 read, and openjpeg doesn't have a way to be told "EOF and
	// data", so we ignore EOFs if any data was read from the stream
	if err == io.EOF && n > 0 {
		err = nil
	}

	if err != nil {
		if err != io.EOF {
			Logger.Errorf("Unable to read from stream %d: %s", id, err)
		}
		return opjMinusOneSizeT
	}

	return C.OPJ_SIZE_T(n)
}

//export opjStreamSkip
//
// opjStreamSkip jumps numBytes ahead in the stream, discarding any data that would be read
func opjStreamSkip(numBytes C.OPJ_OFF_T, id uint64) C.OPJ_SIZE_T {
	var i, ok = lookupImage(id)
	if !ok {
		Logger.Errorf("Unable to find stream ID %d", id)
		return opjMinusOneSizeT
	}
	var _, err = i.streamer.Seek(int64(numBytes), io.SeekCurrent)
	if err != nil {
		Logger.Errorf("Unable to seek %d bytes forward: %s", numBytes, err)
		return opjMinusOneSizeT
	}

	// For some reason, success here seems to be a return value of the number of bytes passed in
	return C.OPJ_SIZE_T(numBytes)
}

//export opjStreamSeek
//
// opjStreamSeek jumps to the absolute position offset in the stream
func opjStreamSeek(offset C.OPJ_OFF_T, id uint64) C.OPJ_BOOL {
	var i, ok = lookupImage(id)
	if !ok {
		Logger.Errorf("Unable to find stream ID %d", id)
		return C.OPJ_FALSE
	}
	var _, err = i.streamer.Seek(int64(offset), io.SeekStart)
	if err != nil {
		Logger.Errorf("Unable to seek to offset %d: %s", offset, err)
		return C.OPJ_FALSE
	}

	return C.OPJ_TRUE
}
