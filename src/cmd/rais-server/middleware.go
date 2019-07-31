package main

import (
	"net/http"
	"rais/src/cmd/rais-server/internal/statusrecorder"
)

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ip = r.RemoteAddr
		var forwarded = r.Header.Get("X-Forwarded-For")
		if forwarded != "" {
			ip = ip + "," + forwarded
		}
		var sr = statusrecorder.New(w)
		next.ServeHTTP(sr, r)
		Logger.Infof("Request: [%s] %s - %d", ip, r.URL, sr.Status)
	})
}
