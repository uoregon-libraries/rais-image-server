package main

import "net/http"

func (s *serverStats) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var json, err = s.Serialize()
	if err != nil {
		http.Error(w, "error generating json: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
