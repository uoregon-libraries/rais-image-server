package main

import (
	"net/http"
	"rais/src/cmd/rais-server/internal/statusrecorder"

	"github.com/uoregon-libraries/gopkg/logger"
)

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ip = r.RemoteAddr
		logger.Infof("%#v", r.Header)
		var forwarded = r.Header.Get("X-Forwarded-For")
		if forwarded != "" {
			ip = ip + "," + forwarded
		}
		var sr = statusrecorder.New(w)
		next.ServeHTTP(sr, r)
		logger.Infof("Request: [%s] %s - %d", ip, r.URL, sr.Status)
	})
}
