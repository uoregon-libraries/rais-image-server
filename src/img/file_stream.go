package img

import (
	"net/url"
	"os"
)

// FileStream simply wraps an os.File to provide streaming functionality
type FileStream struct {
	filepath string
	*os.File
}

// NewFileStream returns a streamer for the given URL's path.  If the URL's
// path doesn't refer to a file on the local filesystem, this will fail in
// stupid ways.
func NewFileStream(path string) (*FileStream, error) {
	var f, err = os.Open(path)
	// Make sure the most common error, at least, will get reported *our* way
	// (e.g., translated to a 404 when this is done via a web request)
	if os.IsNotExist(err) {
		return nil, ErrDoesNotExist
	}
	return &FileStream{filepath: path, File: f}, err
}

// Location returns a "file://" location based on the original path
func (fs *FileStream) Location() *url.URL {
	return &url.URL{Scheme: "file", Path: fs.filepath}
}

// Exist checks the filesystem to be sure the file is there.
//
// This is a very naive check - if stat returns a "doesn't exist" error, we say
// no, but we don't try to decide what other cases may mean, such as lack of
// permissions.  Those may mean the file exists and that isn't for this
// function to handle.
func (fs *FileStream) Exist() bool {
	var _, err = os.Stat(fs.filepath)
	return !os.IsNotExist(err)
}
