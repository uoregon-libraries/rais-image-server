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
var indexT, assetT, adminT *template.Template
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
		fmt.Println("AWS_SECRET_ACCESS_KEY before running the demo.  You can export these directly")
		fmt.Println(`or use a the docker-compose ".env" file.`)
		os.Exit(1)
	}

	readAssets()
	preptemplates()
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
		var id = url.PathEscape(fmt.Sprintf("s3://%s/%s", bucket, key))
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
	adminT = template.Must(template.Must(layout.Clone()).ParseFiles("admin.go.html"))
}

type Data struct {
	Zone      string
	Bucket    string
	KeyID     string
	SecretKey string
}
