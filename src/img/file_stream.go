package img

import (
	"net/url"
	"os"
	"time"
)

// FileStream simply wraps an os.File to provide streaming functionality
type FileStream struct {
	filepath string
	info     os.FileInfo
	*os.File
}

// NewFileStream returns a streamer for the given URL's path.  If the URL's
// path doesn't refer to a file on the local filesystem, this will fail in
// stupid ways.
func NewFileStream(path string) (*FileStream, error) {
	var fs = &FileStream{filepath: path}
	var err error

	fs.info, err = os.Stat(path)

	// Make sure the most common error, at least, will get reported *our* way
	// (e.g., translated to a 404 when this is done via a web request)
	if os.IsNotExist(err) {
		return nil, ErrDoesNotExist
	}
	if err != nil {
		return nil, err
	}

	fs.File, err = os.Open(path)
	return fs, err
}

// Location returns a "file://" location based on the original path
func (fs *FileStream) Location() *url.URL {
	return &url.URL{Scheme: "file", Path: fs.filepath}
}

// Size returns the file's length in bytes
func (fs *FileStream) Size() int64 {
	return fs.info.Size()
}

// ModTime returns when the file was last changed
func (fs *FileStream) ModTime() time.Time {
	return fs.info.ModTime()
}
