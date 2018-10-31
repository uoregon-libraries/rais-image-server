// s3list.go is a slightly modified version of the S3 object list example in
// the AWS repo... mostly because that example didn't work....

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: go run scripts/s3list.go <region> <bucket>\n\n")
		os.Exit(1)
	}

	var conf = &aws.Config{Region: &os.Args[1]}
	var sess, err = session.NewSession(conf)
	if err != nil {
		fmt.Println("Error trying to instantiate a new AWS session: ", err)
		os.Exit(1)
	}
	var svc = s3.New(sess)

	var out *s3.ListObjectsOutput
	out, err = svc.ListObjects(&s3.ListObjectsInput{Bucket: &os.Args[2]})
	if err != nil {
		log.Println("Error trying to list objects: ", err)
		log.Println("Make sure you have your AWS credentials set up in ~/.aws/credentials or exported to environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY")
		os.Exit(1)
	}

	for _, obj := range out.Contents {
		fmt.Println(*obj.Key)
	}
}
