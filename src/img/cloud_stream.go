package img

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

// Environment variables which CloudStream uses to set up S3
const (
	EnvS3Endpoint       = "RAIS_S3_ENDPOINT"
	EnvS3DisableSSL     = "RAIS_S3_DISABLESSL"
	EnvS3ForcePathStyle = "RAIS_S3_FORCEPATHSTYLE"
)

// CloudStream uses gocloud.dev tools to open common types of external streams
type CloudStream struct {
	cleanURL  *url.URL
	bucketURL string
	key       string
	bucket    *blob.Bucket
	size      int64
	modTime   time.Time
	offset    int64
	ctx       context.Context
	r         *blob.Reader
}

// OpenStream returns a CloudStream for the given URL.
//
// We don't allow *anything* except scheme, hostname (bucket), and path in
// streamable URLs.  There's no need for anything else on the local filesystem,
// we have to set up custom values for S3, and we wouldn't want to allow for
// potential security issues by letting literally any data through from an
// Internet request (e.g., if some custom query parameter one day makes an
// operation destructive)
func OpenStream(u *url.URL) (s *CloudStream, err error) {
	// Determine initial data for the bucket and key
	s = new(CloudStream)
	err = s.initialize(u)
	if err != nil {
		return nil, err
	}

	// TODO: let the server pass context in, but *document it clearly* that
	// OpenStream takes a context and that context must remain open until the
	// operation is completed (e.g., http request).
	s.ctx = context.Background()
	s.bucket, err = blob.OpenBucket(s.ctx, s.bucketURL)
	if err != nil {
		return nil, err
	}

	var exists bool
	exists, err = s.bucket.Exists(s.ctx, s.key)
	if err != nil {
		return nil, err
	}

	// We call out a nonexistent blob properly so the server can return a 404
	// rather than a 500
	if !exists {
		return nil, ErrDoesNotExist
	}

	return s, s.getMetadata()
}

// initialize sets up the data based on a given URL, calculating things like
// the bucket URL and key, and storing a "clean" URL in the stream
func (s *CloudStream) initialize(u *url.URL) error {
	s.cleanURL = new(url.URL)
	s.cleanURL.Scheme = u.Scheme
	s.cleanURL.Host = u.Host
	s.cleanURL.Path = u.Path
	s.cleanURL.RawPath = u.RawPath

	var usablePath = s.cleanURL.Path
	if usablePath == "" {
		usablePath = s.cleanURL.RawPath
	}
	if usablePath == "" {
		return errors.New("invalid url path")
	}

	// Easy path first: local files' buckets are the full path, not just the
	// initial path element, and the "key" is just the filename
	if s.cleanURL.Scheme == "file" {
		var dir string
		dir, s.key = path.Split(usablePath)
		s.bucketURL = "file:///" + dir
		return nil
	}

	var pathParts = strings.Split(usablePath, "/")

	// This shouldn't be possible, so it's a full-on panic here
	if len(pathParts) < 2 {
		panic(fmt.Sprintf("img.NewStream: invalid path in URL %q", u))
	}

	s.bucketURL = u.Scheme + "://" + u.Host
	if usablePath[0] == '/' {
		usablePath = usablePath[1:]
	}
	s.key = usablePath
	s.applyEnvironmentConfiguration()

	return nil
}

// applyEnvironmentConfiguration uses any cloud-specific environment settings
// which need to alter the stream's data in some way
func (s *CloudStream) applyEnvironmentConfiguration() {
	// As far as I know, only S3 needs this magic for now, so we short-circuit
	// the function if the scheme isn't S3
	if s.cleanURL.Scheme != "s3" {
		return
	}

	var endpoint = os.Getenv(EnvS3Endpoint)
	var disableSSL = os.Getenv(EnvS3DisableSSL)
	var forcePathStyle = os.Getenv(EnvS3ForcePathStyle)
	var query []string

	if endpoint != "" {
		query = append(query, "endpoint="+endpoint)
	}

	// Allow "t", "T", "true", "True", etc.
	if disableSSL != "" && strings.ToLower(disableSSL)[:1] == "t" {
		query = append(query, "disableSSL=true")
	}

	if forcePathStyle != "" && strings.ToLower(forcePathStyle)[:1] == "t" {
		query = append(query, "s3ForcePathStyle=true")
	}

	if len(query) == 0 {
		return
	}

	s.bucketURL += "?" + strings.Join(query, "&")
}

func (s *CloudStream) getMetadata() error {
	// Store the modtime and size since those actually require a reader to be
	// created and then destroyed just to get that info
	var r, err = s.bucket.NewReader(s.ctx, s.key, nil)
	if err != nil {
		return err
	}
	defer r.Close()

	s.size = r.Size()
	s.modTime = r.ModTime()

	return nil
}

// Location returns the "clean" url for the blob
func (s *CloudStream) Location() *url.URL {
	return s.cleanURL
}

// Size returns the object's length in bytes
func (s *CloudStream) Size() int64 {
	return s.size
}

// ModTime returns when the object was last changed
func (s *CloudStream) ModTime() time.Time {
	return s.modTime
}

// Read implements io.Reader
func (s *CloudStream) Read(buf []byte) (n int, err error) {
	// Create a blob.Reader that is set to our current position
	if s.r == nil {
		s.r, err = s.bucket.NewRangeReader(s.ctx, s.key, s.offset, -1, nil)
	}
	if err != nil {
		return 0, err
	}

	n, err = s.r.Read(buf)
	s.offset += int64(n)
	return n, err
}

// Gently stolen from the cold, dead hands of io.go
var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

// Seek implements io.Seeker.  Since we don't actually move a real file
// pointer, this just stores our internal position for the next Read() call.
//
// Valid whence values:
//
//     - io.SeekStart means relative to the start of the file
//     - io.SeekCurrent means relative to the current offset
//     - io.SeekEnd means relative to the end
//
// Seek returns the new offset relative to the start of the
// file and an error, if any.  If the final offset ends up being below zero, an
// error will be returned and the pointer will remain unchanged.
func (s *CloudStream) Seek(offset int64, whence int) (int64, error) {
	var orig = s.offset

	switch whence {
	default:
		return 0, errWhence
	case io.SeekStart:
		s.offset = offset
	case io.SeekCurrent:
		s.offset += offset
	case io.SeekEnd:
		s.offset = s.size + offset
	}

	if s.offset < 0 {
		s.offset = orig
		return 0, errOffset
	}

	// If our offset changed, the reader is no longer valid and must be closed
	if orig != s.offset {
		s.closeReader()
	}

	return s.offset, nil
}

// Close implements io.Closer and frees the bucket's resources
func (s *CloudStream) Close() error {
	s.closeReader()
	return s.bucket.Close()
}

// closeReader lets us have a one-line close operation only when the reader is
// open.  We remove the reference once we close it to ensure it's not reused
// accidentally.
func (s *CloudStream) closeReader() {
	if s.r != nil {
		s.r.Close()
		s.r = nil
	}
}
