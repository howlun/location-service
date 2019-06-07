package terminal

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/iknowhtml/locationtracker/pkg/common"
	"github.com/iknowhtml/locationtracker/pkg/config"
	"github.com/iknowhtml/locationtracker/pkg/location"
	//keycloak "github.com/mitch-strong/keycloakgo"
)

// HTTPServer holds the necessary structure for our
// HTTP server.
type HTTPServer struct {
	CorsConfig config.CORSConfig
	Addr       string
	AuthServer string
	Wg         *sync.WaitGroup
	Router     *mux.Router
	Server     *http.Server
}

func (u *HTTPServer) New() *HTTPServer {
	log.Printf("Initializing HTTP server: %s\n", u.Addr)

	// initialize Auth Server
	//srvURI := "http://" + u.Addr
	//authSrvURI := "https://" + u.AuthServer
	//keycloak.Init(authSrvURI, srvURI)

	// initialize HTTP router for "fleet" domain
	router := mux.NewRouter()
	router.Schemes("http")
	router.Use(common.MainMiddleware)

	// add Common route
	for _, r := range common.NewRouter() {
		router.
			Methods(r.Method).
			Path(r.Pattern).
			Name(r.Name).
			Handler(r.HandlerFunc)
	}

	// add Location route
	fleetAPI := router.PathPrefix("/api/" + location.PathPrefix).Subrouter()
	fleetAPI.Use(common.RESTResponseMiddleware)
	for _, r := range location.NewRouter() {

		// adding in Authentication middleware
		//authHandler := keycloak.AuthMiddleware(r.HandlerFunc)

		//fleetAPI.Methods(r.Method).Path(r.Pattern).Name(r.Name).Handler(authHandler)
		fleetAPI.Methods(r.Method).Path(r.Pattern).Name(r.Name).Handler(r.HandlerFunc)
	}

	// create CORS middleware
	cors := common.CORSMiddlewareObj{
		AllowedOrigins:     u.CorsConfig.AllowedOrigins,
		AllowedMethods:     u.CorsConfig.AllowedMethods,
		AllowedHeaders:     u.CorsConfig.AllowedHeaders,
		AllowCredentials:   u.CorsConfig.AllowCredentials,
		Debug:              u.CorsConfig.Debug,
		OptionsPassthrough: u.CorsConfig.OptionsPassthrough,
		MaxAge:             u.CorsConfig.MaxAge,
	}

	u.Server = &http.Server{
		Addr: u.Addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      cors.CORSMiddleware(router), // inject CORS middleware to handle OPTIONS header
	}

	log.Printf("HTTP Server initialized: %s\n", u.Addr)

	return u
}

// Process will take the data from channel for processing.
func (u *HTTPServer) Process() {
	log.Printf("Processing data: %s\n", u.Addr)
	log.Printf("Finished processing data: %s\n", u.Addr)
}

// Run starts the HTTP server.
func (u *HTTPServer) Run() {
	log.Printf("Running server: %s\n", u.Addr)
	// close the HTTP server when finished running
	defer u.Close()

	// signal the system to wait for server to finished running before exiting
	defer u.Wg.Done()

	err := u.Server.ListenAndServe()
	if err != nil {
		log.Panicf("HTTP Server failed: %s\n", err.Error())
	}

	//log.Printf("Server stopped: %s\n", u.Addr)
}

// Close ensures that the HTTPServer is shut down gracefully.
func (u *HTTPServer) Close() error {
	log.Printf("Closing server: %s\n", u.Addr)
	return u.Server.Shutdown(context.TODO())
}
