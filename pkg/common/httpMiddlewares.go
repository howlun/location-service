package common

import (
	"log"
	"net/http"
	"time"

	"github.com/rs/cors"
)

func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Println("Setting Content-Type to application/json...")
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func RESTResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.Println("Setting REST Response Wrapper...")
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)

		// Check if response is type of StatusWriter
		sw, ok := w.(*StatusWriter)
		if ok {
			log.Println(sw.ResponseWriter)
		}
	})
}

func MainMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Setting Main HTTP Request and Response Wrapper...")

		start := time.Now()
		sw := StatusWriter{ResponseWriter: w}

		next.ServeHTTP(&sw, r)

		log.Println(HTTPLogEntry{
			Host:            r.Host,
			RemoteAddr:      r.RemoteAddr,
			Method:          r.Method,
			RequestURI:      r.RequestURI,
			Proto:           r.Proto,
			Status:          sw.status,
			ContentLen:      sw.length,
			UserAgent:       r.Header.Get("User-Agent"),
			DurationSeconds: time.Since(start).Seconds(),
		})
	})
}

type CORSMiddlewareObj struct {
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	AllowCredentials   bool
	Debug              bool
	OptionsPassthrough bool
	MaxAge             int32
}

func (cm *CORSMiddlewareObj) CORSMiddleware(next http.Handler) http.Handler {
	log.Printf("Setting CORS settings: %v\n", cm)

	c := cors.New(cors.Options{
		AllowedOrigins:     cm.AllowedOrigins,
		AllowedMethods:     cm.AllowedMethods,
		AllowedHeaders:     cm.AllowedHeaders,
		AllowCredentials:   cm.AllowCredentials,
		Debug:              cm.Debug,
		OptionsPassthrough: cm.OptionsPassthrough,
		MaxAge:             int(cm.MaxAge),
	})

	return c.Handler(next)
}
