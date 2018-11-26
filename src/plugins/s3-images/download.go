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

func (a *asset) fetch() error {
	var conf = &aws.Config{Region: aws.String(s3zone)}
	var sess, err = session.NewSession(conf)
	if err != nil {
		return fmt.Errorf("unable to set up AWS session: %s", err)
	}

	var obj = &s3.GetObjectInput{
		Bucket: aws.String(s3bucket),
		Key:    aws.String(a.key),
	}

	var fullpath = filepath.Dir(a.path)
	err = os.MkdirAll(fullpath, 0755)
	if err != nil {
		return fmt.Errorf("unable to create cached file path %q: %s", fullpath, err)
	}
	var tmpfile = fileutil.NewSafeFile(a.path)

	var dl = s3manager.NewDownloader(sess)
	_, err = dl.Download(tmpfile, obj)
	if err != nil {
		tmpfile.Cancel()
		return fmt.Errorf("unable to download item %q: %s", a.key, err)
	}

	return tmpfile.Close()
}
