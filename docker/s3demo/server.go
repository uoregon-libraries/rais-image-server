package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func serve() {
	http.HandleFunc("/", renderIndex)
	http.HandleFunc("/asset/", renderAsset)
	http.HandleFunc("/api/", renderAPIForm)

	var fileServer = http.FileServer(http.Dir("."))
	http.Handle("/osd/", fileServer)

	log.Println("Listening on port 8080")
	var err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Printf("Error trying to serve http: %s", err)
	}
}

type indexData struct {
	Bucket string
	Assets []asset
}

func renderIndex(w http.ResponseWriter, req *http.Request) {
	var path = req.URL.Path
	if path != "/" {
		http.NotFound(w, req)
		return
	}
	var data = indexData{Assets: s3assets, Bucket: bucketName}
	var err = indexT.Execute(w, data)
	if err != nil {
		log.Printf("Unable to serve index: %s", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func findAssetKey(req *http.Request) string {
	var p = req.URL.RawPath
	if p == "" {
		p = req.URL.Path
	}
	var parts = strings.Split(p, "/")
	if len(parts) < 3 {
		log.Printf("Invalid path parts %#v", parts)
		return ""
	}

	return strings.Join(parts[2:], "/")
}

func findAsset(key string) asset {
	for _, a2 := range s3assets {
		if a2.Key == key {
			return a2
		}
	}

	return emptyAsset
}

func renderAsset(w http.ResponseWriter, req *http.Request) {
	var key = findAssetKey(req)
	if key == "" {
		http.Error(w, "invalid asset request", http.StatusBadRequest)
		return
	}

	var a = findAsset(key)
	if a == emptyAsset {
		log.Printf("Invalid asset key %q", key)
		http.Error(w, fmt.Sprintf("Asset %q doesn't exist", key), http.StatusNotFound)
		return
	}

	var err = assetT.Execute(w, map[string]interface{}{"Asset": a})
	if err != nil {
		log.Printf("Unable to serve asset page: %s", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func renderAPIForm(w http.ResponseWriter, req *http.Request) {
	var err = adminT.Execute(w, nil)
	if err != nil {
		log.Printf("Unable to serve admin page: %s", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}
