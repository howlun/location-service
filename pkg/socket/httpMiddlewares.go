package socket

import (
	"log"
	"net/http"

	"github.com/iknowhtml/locationtracker/pkg/common"
)

func WebSocketMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Println("Setting WebSocket Response Wrapper...")
		//w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func MainMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Setting Main Socket Request and Response Wrapper...")

		log.Println(common.HTTPLogEntry{
			Host:       r.Host,
			RemoteAddr: r.RemoteAddr,
			Method:     r.Method,
			RequestURI: r.RequestURI,
			Proto:      r.Proto,
			UserAgent:  r.Header.Get("User-Agent"),
		})

		next.ServeHTTP(w, r)
	})
}
