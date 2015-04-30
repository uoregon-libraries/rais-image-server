package main

import (
	"net/http"
	"os"
	"time"
)

func sendHeaders(w http.ResponseWriter, req *http.Request, filepath string) error {
	info, err := os.Stat(filepath)
	if err != nil {
		http.Error(w, "Unable to access file", 404)
		return err
	}

	// Set headers
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Last-Modified", info.ModTime().Format(time.RFC1123))

	// Check for forced download parameter
	query := req.URL.Query()
	if query["download"] != nil {
		w.Header().Set("Content-Disposition", "attachment")
	}

	return nil
}
