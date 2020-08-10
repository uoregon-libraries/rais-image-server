package main

import (
	"net/http"
	"rais/src/img"
	"time"
)

func sendHeaders(w http.ResponseWriter, req *http.Request, res *img.Resource) error {
	// Set headers
	w.Header().Set("Last-Modified", res.Streamer().ModTime().Format(time.RFC1123))
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check for forced download parameter
	query := req.URL.Query()
	if query["download"] != nil {
		w.Header().Set("Content-Disposition", "attachment")
	}

	return nil
}
