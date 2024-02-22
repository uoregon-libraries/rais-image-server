package img

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/uoregon-libraries/gopkg/assert"
)

func TestS3URL(t *testing.T) {
	assert.NilError(os.Setenv(EnvS3Endpoint, ""), "Setenv succeeded", t)
	assert.NilError(os.Setenv(EnvS3DisableSSL, ""), "Setenv succeeded", t)
	assert.NilError(os.Setenv(EnvS3ForcePathStyle, ""), "Setenv succeeded", t)

	var u, _ = url.Parse("s3://mybucket/path/to/asset.jp2")
	var s = new(CloudStream)
	var err = s.initialize(u)
	if err != nil {
		t.Errorf("Unable to initialize %#v: %s", u, err)
	}

	var expected = "s3://mybucket"
	if s.bucketURL != expected {
		t.Errorf("expected bucket to be %q, got %q", expected, s.bucketURL)
	}

	expected = "path/to/asset.jp2"
	if s.key != expected {
		t.Errorf("expected key to be %q, got %q", expected, s.key)
	}
}

func TestS3CustomURL(t *testing.T) {
	assert.NilError(os.Setenv(EnvS3Endpoint, "minio:9000"), "Setenv succeeded", t)
	assert.NilError(os.Setenv(EnvS3DisableSSL, "true"), "Setenv succeeded", t)
	assert.NilError(os.Setenv(EnvS3ForcePathStyle, "false"), "Setenv succeeded", t)

	var u, _ = url.Parse("s3://mybucket/path/to/asset.jp2")
	var s = new(CloudStream)
	var err = s.initialize(u)
	if err != nil {
		t.Errorf("Unable to initialize %#v: %s", u, err)
	}

	var expected = "s3://mybucket?endpoint=minio:9000&disableSSL=true"
	if s.bucketURL != expected {
		t.Errorf("expected bucket to be %q, got %q", expected, s.bucketURL)
	}
}

func openFile(testPath string) (realFile *os.File, cloudFile *CloudStream, info os.FileInfo) {
	var err error
	realFile, err = os.Open(testPath)
	if err != nil {
		panic(fmt.Sprintf("os.Open(%q) error: %s", testPath, err))
	}

	info, _ = realFile.Stat()
	if err != nil {
		_ = realFile.Close()
		panic(fmt.Sprintf("realFile.Stat() error: %s", err))
	}

	var u, _ = url.Parse("file://" + testPath)
	cloudFile, err = OpenStream(u)
	if err != nil {
		panic(fmt.Sprintf("OpenStream(%q) error: %s", "file://"+testPath, err))
	}

	return realFile, cloudFile, info
}

func testRead(a, b io.Reader, bufsize int, t *testing.T) {
	var err error
	var aDat = make([]byte, bufsize)
	var bDat = make([]byte, bufsize)
	var aN, bN int

	aN, err = a.Read(aDat)
	if err != nil {
		panic(fmt.Sprintf("error reading a: %s", err))
	}
	bN, err = b.Read(bDat)
	if err != nil {
		panic(fmt.Sprintf("error reading b: %s", err))
	}
	if aN != bN {
		t.Errorf("a read %d; b read %d", aN, bN)
	}
	if string(aDat) != string(bDat) {
		t.Errorf("aDat %q didn't match bDat %q", aDat, bDat)
	}
}

func testSeek(a, b io.Seeker, offset int64, whence int, t *testing.T) {
	var err error
	var aN, bN int64
	aN, err = a.Seek(offset, whence)
	if err != nil {
		panic(fmt.Sprintf("error seeking a: %s", err))
	}
	bN, err = b.Seek(offset, whence)
	if err != nil {
		panic(fmt.Sprintf("error seeking b: %s", err))
	}

	if aN != bN {
		t.Errorf("seek(%d, %d) returns %d for a but %d for b", offset, whence, aN, bN)
	}
}

func TestRandomAccess(t *testing.T) {
	var dir, err = os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error: %s", err)
	}
	var testPath = path.Join(dir, "../../docker/images/jp2tests/sn00063609-19091231.jp2")

	var realFile, cloudFile, info = openFile(testPath)
	if info.Size() != cloudFile.Size() {
		t.Errorf("realFile size %d; cloudFile size %d", info.Size(), cloudFile.Size())
	}

	testRead(realFile, cloudFile, 8192, t)
	testRead(realFile, cloudFile, 10240, t)
	testSeek(realFile, cloudFile, 50000, 0, t)
	testRead(realFile, cloudFile, 10240, t)
	testSeek(realFile, cloudFile, 50000, 1, t)
	testRead(realFile, cloudFile, 10240, t)
}
