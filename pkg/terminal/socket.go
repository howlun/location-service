package terminal

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/iknowhtml/locationtracker/pkg/common"
	"github.com/iknowhtml/locationtracker/pkg/socket"
)

// SocketServer holds the necessary structure for our
// Socket server.
type SocketServer struct {
	Addr   string
	Wg     *sync.WaitGroup
	Server *http.Server
}

func (s *SocketServer) New() *SocketServer {
	log.Printf("Initializing Socket server: %s\n", s.Addr)

	// initialize Socket router
	router := mux.NewRouter()
	router.Schemes("http")
	router.Use(socket.MainMiddleware)

	// add Common route
	for _, r := range common.NewWSRouter() {
		router.
			Methods(r.Method).
			Path(r.Pattern).
			Name(r.Name).
			Handler(r.HandlerFunc)
	}

	// add Location route
	socketRouter := router.PathPrefix("/socket/").Subrouter()
	socketRouter.Use(socket.WebSocketMiddleware)
	for _, r := range socket.NewWSRouter() {

		// adding in Authentication middleware
		//authHandler := keycloak.AuthMiddleware(r.HandlerFunc)

		socketRouter.Path(r.Pattern).Name(r.Name).Handler(r.HandlerFunc)
	}

	s.Server = &http.Server{
		Addr: s.Addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		//WriteTimeout: writeWait,
		//ReadTimeout:  readWait,
		//IdleTimeout:  timeoutPeriod,
		Handler: router,
	}

	log.Printf("Socket Server initialized: %s\n", s.Addr)

	return s
}

// Process will take the data from channel for processing.
func (s *SocketServer) Process() {
	log.Printf("Processing data: %s\n", s.Addr)
	log.Printf("Finished processing data: %s\n", s.Addr)
}

// Run starts the Socket server.
func (s *SocketServer) Run() {
	log.Printf("Running server: %s\n", s.Addr)
	// close the Socket server when finished running
	defer s.Close()

	// signal the system to wait for server to finished running before exiting
	defer s.Wg.Done()

	err := s.Server.ListenAndServe()
	if err != nil {
		log.Panicf("Socket Server failed: %s\n", err.Error())
	}
}

// Close ensures that the SocketServer is shut down gracefully.
func (s *SocketServer) Close() error {
	log.Printf("Closing server: %s\n", s.Addr)
	return s.Server.Shutdown(context.TODO())
}
