package utils

import (
	"log"
	"net/http"
)

func HttpRouterLogging(logger *log.Logger, next http.Handler) http.Handler {
	// return func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
		next.ServeHTTP(w, r)
	})
	// }
}
