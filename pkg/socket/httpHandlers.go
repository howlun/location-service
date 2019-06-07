package socket

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/iknowhtml/locationtracker/pkg/common"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Socket/WebSocketHandler: Handling websocket request...\n")
	if r.Method != "GET" {
		common.HandleMethodNotAllowedResponse(w, "")
		return
	}

	// Start hub
	hub := newHub()
	go hub.run()

	ServeWs(hub, w, r, 0)
}

func WSDriverStatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Socket/WSDriverStatusHandler: Handling Driver Status request...\n")
	if r.Method != "GET" {
		common.HandleMethodNotAllowedResponse(w, "")
		return
	}

	vars := mux.Vars(r)

	var driverID int
	if vars["id"] != "" && vars["id"] != "0" {
		driverID, _ = strconv.Atoi(vars["id"])
	} else {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Driver ID is missing, but required")
		return
	}

	// Start hub
	hub := newHub()
	go hub.run()

	ServeWs(hub, w, r, int32(driverID))
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, driverID int32) {
	log.Printf("Socket/ServeWs: Upgrading connection...\n")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Socket/ServeWs: Error upgrading connection: %v\n", err)
		return
	}
	log.Printf("Socket/ServeWs: New connection: %d\n", driverID)
	client := &Client{clientId: driverID, hub: hub, conn: conn, send: make(chan []byte, 256), status: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WriteMessage()
	go client.ReadMessage()
}
