// Package main, along with the various *.go.html files, demonstrates a very
// simple (and ugly) asset server that reads all S3 assets in a given region
// and bucket, and serves up HTML pages which point to a IIIF server (RAIS, of
// course) for thumbnails and full-image views.
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"strings"
	tt "text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type asset struct {
	Key    string
	IIIFID string
	Title  string
}

var emptyAsset asset

var s3assets []asset
var indexT, assetT *template.Template
var zone, bucket string
var keyID, secretKey string

func env(key string) string {
	for _, kvpair := range os.Environ() {
		var parts = strings.SplitN(kvpair, "=", 2)
		if parts[0] == key {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

func main() {
	zone = env("RAIS_S3ZONE")
	bucket = env("RAIS_S3BUCKET")
	keyID = env("AWS_ACCESS_KEY_ID")
	secretKey = env("AWS_SECRET_ACCESS_KEY")

	if zone == "" || bucket == "" || keyID == "" || secretKey == "" {
		fmt.Println("You must set env vars RAIS_S3BUCKET, RAIS_S3ZONE, AWS_ACCESS_KEY_ID, and")
		fmt.Println("AWS_SECRET_ACCESS_KEY before running the demo")
		os.Exit(1)
	}

	readAssets()
	preptemplates()
	writeDockerCompose()
	serve()
}

func readAssets() {
	var conf = &aws.Config{Region: &zone}
	var sess, err = session.NewSession(conf)
	if err != nil {
		log.Println("Error trying to instantiate a new AWS session: ", err)
		os.Exit(1)
	}
	var svc = s3.New(sess)

	var out *s3.ListObjectsOutput
	out, err = svc.ListObjects(&s3.ListObjectsInput{Bucket: &bucket})
	if err != nil {
		log.Println("Error trying to list objects: ", err)
		os.Exit(1)
	}

	for _, obj := range out.Contents {
		var key = *obj.Key
		var id = "s3:" + url.PathEscape(key)
		s3assets = append(s3assets, asset{Title: key, Key: key, IIIFID: id})
	}
	log.Printf("Indexed %d assets", len(s3assets))
}

func preptemplates() {
	var _, err = os.Stat("./layout.go.html")
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Unable to load HTML layout: file does not exist.  Make sure you run the demo from the docker/s3demo folder.")
		} else {
			log.Printf("Error trying to open layout: %s", err)
		}
		os.Exit(1)
	}

	var root = template.New("layout")
	var layout = template.Must(root.ParseFiles("layout.go.html"))
	indexT = template.Must(template.Must(layout.Clone()).ParseFiles("index.go.html"))
	assetT = template.Must(template.Must(layout.Clone()).ParseFiles("asset.go.html"))
}

type Data struct {
	Zone      string
	Bucket    string
	KeyID     string
	SecretKey string
}

func writeDockerCompose() {
	var t = tt.Must(tt.ParseFiles("./docker-compose.template.yml"))
	var data = Data{
		Zone:      zone,
		Bucket:    bucket,
		KeyID:     keyID,
		SecretKey: secretKey,
	}

	var f, err = os.Create("./docker-compose.yml")
	if err != nil {
		log.Printf("Unable to create docker-compose.yml: %s", err)
		os.Exit(1)
	}

	f.WriteString("# This is a generated file: **do not modify**\n")
	f.WriteString("#\n")
	f.WriteString("# If you wish to alter this RAIS S3 demo at the docker level, build a\n")
	f.WriteString("# docker-compose.override.yml file.\n")
	err = t.Execute(f, data)
	if err != nil {
		log.Printf("Unable to build docker-compose.yml file: %s", err)
		f.Close()
		os.Exit(1)
	}

	f.Close()
}
