package terminal

import "log"

// Server defines the minimum contract out TCP and UDP server implementations must satisfy
type ServerHandler interface {
	Run()
	Process()
	Close() error
}

func NewServer(handler ServerHandler) ServerHandler {
	log.Printf("Creating new server...\n")
	if obj, ok := handler.(*UDPServer); ok == true {
		return obj.New()
	} else if obj, ok := handler.(*HTTPServer); ok == true {
		return obj.New()
	} else if obj, ok := handler.(*SocketServer); ok == true {
		return obj.New()
	} else {
		log.Fatalln("Factory failed to create server: unknown type")
	}
	return nil
}
