// Package main, along with the various *.go.html files, demonstrates a very
// simple (and ugly) asset server that reads all S3 assets in a given region
// and bucket, and serves up HTML pages which point to a IIIF server (RAIS, of
// course) for thumbnails and full-image views.
package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/s3blob"
)

type asset struct {
	Key    string
	IIIFID string
	Title  string
}

var emptyAsset asset

var s3assets []asset
var indexT, assetT, adminT *template.Template
var s3url, zone, bucketName string
var keyID, secretKey string

func main() {
	bucketName = os.Getenv("RAIS_S3_DEMO_BUCKET")
	keyID = os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	if bucketName == "" || keyID == "" || secretKey == "" {
		fmt.Println("You must set env vars RAIS_S3_DEMO_BUCKET, AWS_ACCESS_KEY_ID, and")
		fmt.Println("AWS_SECRET_ACCESS_KEY before running the demo.  You can export these directly")
		fmt.Println(`or use the docker-compose ".env" file.`)
		os.Exit(1)
	}

	readAssets()
	preptemplates()
	serve()
}

func readAssets() {
	// set up the hard-coded newspaper asset first
	s3assets = append(s3assets, asset{Title: "Local File", Key: "news", IIIFID: "news.jp2"})

	var ctx = context.Background()
	var bucketURL = "s3://" + bucketName + getBucketURLQuery()
	var bucket, err = blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		log.Fatalf("Unable to open S3 bucket %q: %s", bucketURL, err)
	}
	var iter = bucket.List(nil)
	var obj *blob.ListObject
	for {
		obj, err = iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error trying to list assets: %s", err)
		}

		var key = obj.Key
		var id = url.PathEscape(fmt.Sprintf("s3://%s/%s", bucketName, key))
		s3assets = append(s3assets, asset{Title: key, Key: key, IIIFID: id})
	}

	log.Printf("Indexed %d assets", len(s3assets))
}

// Environment variables copied from img.CloudStream
const (
	EnvS3Endpoint       = "RAIS_S3_ENDPOINT"
	EnvS3DisableSSL     = "RAIS_S3_DISABLESSL"
	EnvS3ForcePathStyle = "RAIS_S3_FORCEPATHSTYLE"
)

func getBucketURLQuery() string {
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
		return ""
	}

	return "?" + strings.Join(query, "&")
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
