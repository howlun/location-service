package socket

import (
	"log"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iknowhtml/locationtracker/pkg/common"
	"github.com/iknowhtml/locationtracker/pkg/location"
	"github.com/iknowhtml/locationtracker/pkg/message"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Hub ID
	ID string

	// Registered clients.
	clients map[int32]*Client

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

var h *Hub
var once sync.Once

// Initialize and create Hub instance
func newHub() *Hub {
	log.Printf("Socket/newHub: Retrieving Hub...\n")

	once.Do(func() {
		log.Printf("Socket/newHub: Creating new Hub...\n")
		hubId := common.GenUlid()

		h = &Hub{
			ID:         hubId,
			broadcast:  make(chan []byte),
			register:   make(chan *Client),
			unregister: make(chan *Client),
			clients:    make(map[int32]*Client),
		}
	})

	return h
}

// Run the Hub instance to start registering/unregistering clients and handling of incoming messages
func (h *Hub) run() {
	log.Printf("Socket/run: Running Hub: %s\n", h.ID)
	for {
		select {
		case client := <-h.register:
			log.Printf("Socket/run: Registering new client: %v\n", client)
			h.clients[client.clientId] = client

			// start a new goroutine to pushing status every x seconds
			go h.startPushStatus(client, writePeriod)

		case client := <-h.unregister:
			log.Printf("Socket/run: Un-Registering client: %v\n", client)
			if _, ok := h.clients[client.clientId]; ok {
				// remove client
				log.Printf("Socket/run:Remove client: %v\n", client)
				delete(h.clients, client.clientId)
				// close client send channel
				log.Printf("Socket/run:Close client send channel: %v\n", client)
				close(client.send)
				log.Printf("Socket/run:Close client status channel: %v\n", client)
				close(client.status)
			}
		case message := <-h.broadcast:
			log.Printf("Socket/run: Receiving message from broadcost channel: %s\n", string(message))
			for _, client := range h.clients {

				select {
				case client.send <- message:
					log.Printf("Socket/run: Client# %d: Pushing data to Send channel\n", client.clientId)
				default:
					log.Printf("Socket/run: Failed to push message to client send channel: %v\n", client)
					log.Printf("Socket/run:Close client send channel: %v\n", client)
					close(client.send)
					log.Printf("Socket/run:Close client status channel: %v\n", client)
					close(client.status)
					log.Printf("Socket/run:Remove client: %v\n", client)
					delete(h.clients, client.clientId)
				}
			}
		}
	}
	log.Printf("Socket/run: Hub ended...\n")
}

func (h *Hub) startPushStatus(c *Client, pushWait time.Duration) {
	log.Printf("Socket/StartPushStatus: Client# %d: Starting to write message...\n", c.clientId)

	// Get Driver ID
	locationController := new(location.LocationController)
	err := locationController.Init()
	if err != nil {
		log.Printf("Socket/StartPushStatus: Error: %v\n", err.Error())

		// TODO send notification to frontend
		return
	}

	writeTicker := time.NewTicker(pushWait)
	defer func() {
		// recover from panic caused by writing to a closed channel
		if r := recover(); r != nil {
			log.Printf("Socket/StartPushStatus: Client# %d: error writing on channel: %v\n", c.clientId, r)
		}

		writeTicker.Stop()
		c.conn.Close()
		log.Printf("Socket/StartPushStatus: Client# %d: Connection closed\n", c.clientId)
	}()

	// while ticker still running and client still connected
	for {
		select {
		case t := <-writeTicker.C:
			log.Printf("Socket/StartPushStatus: Client# %d: Writing status... %s\n", c.clientId, t.String())
			res, err := locationController.GetDriverStatus(c.clientId)
			if err != nil || !res.Ok {
				log.Printf("Socket/StartPushStatus: Error: %v\n", err.Error())

				// TODO send notification to frontend
				return
			}
			if res.Ok != true {
				log.Printf("Socket/StartPushStatus: Failed to get data: %v\n", res.Error)

				// TODO send notification to frontend
				return
			}

			packet := createPacket(
				res.Fields.DriverID,
				res.Fields.ProviderID,
				res.Object.Coordinates[1],
				res.Object.Coordinates[0],
				res.Fields.Status,
				res.Fields.LastUpdatedTimestamp,
				res.Fields.JobID)
			data, err := proto.Marshal(packet)
			if err != nil {
				log.Panicf("Socket/StartPushStatus: Client# %d: marshalling error: %v\n", c.clientId, err)
			}

			// push status into Status channel
			select {
			case c.status <- data:
				log.Printf("Socket/StartPushStatus: Client# %d: Pushing data to Status channel\n", c.clientId)
			default:
				log.Printf("Socket/StartPushStatus: Failed to push message to client send channel: %v\n", c)
				log.Printf("Socket/StartPushStatus:Close client send channel: %v\n", c)
				close(c.send)
				log.Printf("Socket/StartPushStatus:Close client status channel: %v\n", c)
				close(c.status)
				log.Printf("Socket/StartPushStatus:Remove client: %v\n", c)
				delete(h.clients, c.clientId)
			}
		}
	}

}

func createPacket(
	driverID int32,
	providerID int32,
	lat float32,
	lng float32,
	status location.DriverStatus,
	timestamp int64,
	jobID int32) *message.DriverStatusPoll {
	log.Printf("Socket/createPacket: Client# %d: Creating Status Packet...\n", driverID)
	// create test packet
	packet := message.DriverStatusPoll{
		Fleet:      string(location.Object_Collection_Fleet),
		DriverId:   driverID,
		ProviderId: providerID,
		Lat:        lat,
		Lng:        lng,
		//Status:     message.DriverStatusPoll_BUSY,
		Timestamp: timestamp,
		JobId:     int64(jobID),
	}

	switch status {
	case location.DriverStatus_AVAILABLE:
		packet.Status = message.DriverStatusPoll_AVAILABLE
	case location.DriverStatus_BUSY:
		packet.Status = message.DriverStatusPoll_BUSY
	case location.DriverStatus_NOTAVAILABLE:
		packet.Status = message.DriverStatusPoll_NOTAVAILABLE
	}

	return &packet
}
