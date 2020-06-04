package img

import (
	"net/url"
	"os"
	"testing"
)

func TestS3URL(t *testing.T) {
	os.Setenv(EnvS3Endpoint, "")
	os.Setenv(EnvS3DisableSSL, "")
	os.Setenv(EnvS3ForcePathStyle, "")

	var u, _ = url.Parse("s3://mybucket/path/to/asset.jp2")
	var s = new(CloudStream)
	s.initialize(u)

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
	os.Setenv(EnvS3Endpoint, "minio:9000")
	os.Setenv(EnvS3DisableSSL, "true")
	os.Setenv(EnvS3ForcePathStyle, "false")

	var u, _ = url.Parse("s3://mybucket/path/to/asset.jp2")
	var s = new(CloudStream)
	s.initialize(u)

	var expected = "s3://mybucket?endpoint=minio:9000&disableSSL=true"
	if s.bucketURL != expected {
		t.Errorf("expected bucket to be %q, got %q", expected, s.bucketURL)
	}
}
