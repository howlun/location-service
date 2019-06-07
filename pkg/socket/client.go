package socket

import (
	"bytes"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	clientId int32

	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Buffered channel of status message
	status chan []byte
}

// ReadMessage pull messages from the websocket connection to the hub.
//
// The application runs ReadMessage in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadMessage() {
	log.Printf("Socket/ReadMessage: Client# %d: Starting to read message...\n", c.clientId)
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		log.Printf("Socket/ReadMessage: Client# %d: Connection closed...\n", c.clientId)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		log.Printf("Socket/ReadMessage: Client# %d: Receiving pong message...\n", c.clientId)
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		msgType, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Socket/ReadMessage: Client# %d: Error: %s\n", c.clientId, err.Error())
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Socket/ReadMessage: Client# %d: Unexpected Close Error: %v", c.clientId, err)
			} else {
				log.Printf("Socket/ReadMessage: Client# %d: Error: %v\n", c.clientId, err)
			}
			break
		}

		// handling different message type
		switch msgType {
		case websocket.TextMessage:
			// TextMessage denotes a text data message. The text message payload is
			// interpreted as UTF-8 encoded text data.
			log.Printf("Socket/ReadMessage: Client# %d: Text Message Received: %s\n", c.clientId, string(message))
			message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

			// Send to broadcast channel to send messages to all client's send channel
			//c.hub.broadcast <- message
		case websocket.BinaryMessage:
			// BinaryMessage denotes a binary data message.
			log.Printf("Socket/ReadMessage: Client# %d: Binary Message Received\n", c.clientId)
		case websocket.CloseMessage:
			// CloseMessage denotes a close control message. The optional message
			// payload contains a numeric code and text. Use the FormatCloseMessage
			// function to format a close message payload.
			log.Printf("Socket/ReadMessage: Client# %d: Close Message Received\n", c.clientId)
		case websocket.PingMessage:
			// PingMessage denotes a ping control message. The optional message payload
			// is UTF-8 encoded text.
			log.Printf("Socket/ReadMessage: Client# %d: Ping Message Received\n", c.clientId)
		case websocket.PongMessage:
			// PongMessage denotes a pong control message. The optional message payload
			// is UTF-8 encoded text.
			log.Printf("Socket/ReadMessage: Client# %d: Pong Message Received\n", c.clientId)
		default:
			log.Printf("Socket/ReadMessage: Client# %d: Unknown Message Received\n", c.clientId)
		}
	}
}

// WriteMessage push messages from the hub to the websocket connection.
//
// A goroutine running WriteMessage is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WriteMessage() {
	log.Printf("Socket/WriteMessage: Client# %d: Starting to write message...\n", c.clientId)
	ticker := time.NewTicker(pingPeriod)
	//writeTicker := time.NewTicker(writePeriod)
	defer func() {
		ticker.Stop()
		//writeTicker.Stop()
		c.conn.Close()
		log.Printf("Socket/WriteMessage: Client# %d: Connection closed\n", c.clientId)
	}()

	for {
		select {
		case msg, ok := <-c.send:
			log.Printf("Socket/WriteMessage: Client# %d: Reading from Send channel...\n", c.clientId)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				log.Printf("Socket/WriteMessage: Client# %d: No message from send channel, closing connection...\n", c.clientId)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Printf("Socket/WriteMessage: Client# %d: Writing message: %s\n", c.clientId, msg)
			if err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				log.Printf("Socket/WriteMessage: Client# %d: Error writing message: %v\n", c.clientId, err)
				return
			}
		case status, ok := <-c.status:
			log.Printf("Socket/WriteMessage: Client# %d: Reading from Status channel...\n", c.clientId)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				log.Printf("Socket/WriteMessage: Client# %d: No message from Status channel, closing connection...\n", c.clientId)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Printf("Socket/WriteMessage: Client# %d: Writing status...\n", c.clientId)
			if err := c.conn.WriteMessage(websocket.BinaryMessage, status); err != nil {
				log.Printf("Socket/WriteMessage: Client# %d: Error writing message: %v\n", c.clientId, err)
				return
			}
			/*
				case wt := <-writeTicker.C:
					// write status back to client
					log.Printf("Socket/WriteMessage: Client# %d: Sending status message...\n", c.clientId)
					c.conn.SetWriteDeadline(time.Now().Add(writeWait))
					packet := CreatePacket(c.clientId, wt.Nanosecond())
					data, err := proto.Marshal(packet)
					if err != nil {
						log.Panicf("Socket/WriteMessage: Client# %d: marshalling error: ", c.clientId, err)
					}

					if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
						log.Printf("Socket/WriteMessage: Client# %d: Error sending status message: %v\n", c.clientId, err)
						return
					}
					log.Printf("Socket/WriteMessage: Client# %d: Status message sent: %d\n", c.clientId, len(data))
			*/
		case <-ticker.C:
			// send ping message
			log.Printf("Socket/WriteMessage: Client# %d: Sending ping message...\n", c.clientId)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Socket/WriteMessaNanosecondnt# %d: Error sending ping message: %v\n", c.clientId, err)
				return
			}
		}
	}
}

/*
func CreatePacket(driverId int32, randJobID int) *message.DriverStatusPoll {
	log.Printf("Socket/CreatePacket: Client# %d: Creating Status Packet...\n", driverId)
	// create test packet
	packet := message.DriverStatusPoll{
		Fleet:      "towing",
		DriverId:   driverId,
		ProviderId: 999,
		Lat:        21.045247,
		Lng:        105.845268,
		Status:     message.DriverStatusPoll_BUSY,
		Timestamp:  time.Now().Unix(),
		JobId:      int64(randJobID),
	}
	return &packet
}
*/
