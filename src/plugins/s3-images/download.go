package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/uoregon-libraries/gopkg/fileutil"
)

func (a *asset) setupTempFile() (*fileutil.SafeFile, error) {
	var parentDir = filepath.Dir(a.path)
	var err = os.MkdirAll(parentDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("unable to create cached file path %q: %s", parentDir, err)
	}

	return fileutil.NewSafeFile(a.path), nil
}

func fetchS3(a *asset) error {
	var conf = &aws.Config{
		Region:           aws.String(s3zone),
		Endpoint:         aws.String(s3endpoint),
		S3ForcePathStyle: aws.Bool(true),
	}
	var sess, err = session.NewSession(conf)
	if err != nil {
		return fmt.Errorf("unable to set up AWS session: %s", err)
	}

	var obj = &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(a.key),
	}

	var tmpfile *fileutil.SafeFile
	tmpfile, err = a.setupTempFile()
	if err != nil {
		return err
	}

	var dl = s3manager.NewDownloader(sess)
	_, err = dl.Download(tmpfile, obj)
	if err != nil {
		tmpfile.Cancel()
		return fmt.Errorf("unable to download item %q: %s", a.key, err)
	}

	return tmpfile.Close()
}

func fetchNil(a *asset) error {
	var tmpfile, err = a.setupTempFile()
	if err != nil {
		return err
	}
	return tmpfile.Close()
}
