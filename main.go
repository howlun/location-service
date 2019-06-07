package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iknowhtml/locationtracker/pkg/common"
	"github.com/iknowhtml/locationtracker/pkg/config"
	"github.com/iknowhtml/locationtracker/pkg/message"
	"github.com/iknowhtml/locationtracker/pkg/terminal"

	"github.com/gorilla/websocket"
)

var (
	mode      = flag.String("m", "http", "mode: client or udp or http")
	sentcount = flag.String("c", "1", "sent count: (positive number, only use for UDP client)")
	cid       = flag.String("cid", "1", "client ID: (positive number, only use for socket client)")
	env       = flag.String("e", string(common.EnvType_Dev), "server environment: (dev or prod)")
)

func main() {
	log.Printf("Entering Main\n")
	flag.Parse()

	log.Printf("System running in %s environment...\n", *env)

	// load system configuration based on environment, singleton pattern
	configuration, err := config.GetInstance(*env)
	if configuration == nil {
		log.Panicf("Failed to load configuration: %s\n", err.Error())
	}

	var c_wg, s_wg sync.WaitGroup
	ch := make(chan message.DriverStatusPoll)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	switch *mode {
	case "http":
		// add 1 goroutine to waitgroup
		// 1 - wait for server to finished running
		s_wg.Add(1)

		//server := new(terminal.HTTPServer)
		//server.Init(*port, &s_wg)
		hsh := &terminal.HTTPServer{
			CorsConfig: configuration.Corsconfig,
			Addr:       configuration.Httpserver.Addr,
			AuthServer: configuration.Authserver.RemoteAddr,
			Wg:         &s_wg}
		server := terminal.NewServer(hsh)

		go server.Run()

		// wait for all goroutines to finished
		// 1 - wait for server to finished running
		s_wg.Wait()

		<-stop

		if err := server.Close(); err != nil {
			log.Fatal(err)
		}

	case "socket":
		// add 1 goroutine to waitgroup
		// 1 - wait for server to finished running
		s_wg.Add(1)

		//server := new(terminal.SocketServer)
		//server.Init(*port, &s_wg)
		ssh := &terminal.SocketServer{
			Addr: configuration.Socketserver.Addr,
			Wg:   &s_wg}
		server := terminal.NewServer(ssh)

		go server.Run()

		// wait for all goroutines to finished
		// 1 - wait for server to finished running
		s_wg.Wait()

		<-stop

		if err := server.Close(); err != nil {
			log.Fatal(err)
		}

	case "udp":
		// add 1 goroutine to waitgroup
		// 1 - wait for server to finished running
		s_wg.Add(1)

		//server := new(terminal.UDPServer)
		//server.Init(*port, &s_wg, ch)
		ush := &terminal.UDPServer{Addr: configuration.Udpserver.Addr, Wg: &s_wg, Ch: ch}
		server := terminal.NewServer(ush)

		// start a separate data processing thread first to read data from channel (to prevent a blocking channel)
		go server.Process()

		// start a separate server and data reading thread to read UDP data and store to channel
		go server.Run()

		// wait for all goroutines to finished
		// 1 - wait for server to finished running
		s_wg.Wait()

		<-stop

		if err := server.Close(); err != nil {
			log.Fatal(err)
		}

	case "udpclient":
		// testing client to send data to UDP Server
		client := new(terminal.UDPClient)
		client.Init(configuration.Udpserver.Addr, &c_wg)

		i, _ := strconv.Atoi(*sentcount)
		client.Run(i)
	case "socketclient":
		// add 2 goroutine to waitgroup
		// 1 - wait for reading msg
		// 2 - wait for writting msg
		s_wg.Add(2)

		//u := url.URL{Scheme: "ws", Host: configuration.Socketserver.Addr, Path: "/socket/handle"}
		u := url.URL{Scheme: "ws", Host: configuration.Socketserver.Addr, Path: "/socket/driver/" + *cid + "/status"}
		log.Printf("connecting to %s", u.String())

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		/*
			c.SetPingHandler(func(string) error {
				log.Printf("main/main: Receiving ping message...\n")
				c.SetWriteDeadline(time.Now().Add(60 * time.Second))
				err := c.WriteMessage(websocket.PongMessage, nil)
				if err != nil {
					log.Println("pong write:", err)
					return nil
				}
				return nil
			})
		*/
		go func(s_wg *sync.WaitGroup) {
			defer func() {
				log.Printf("Waitgroup done...\n")
				s_wg.Done()
			}()

			for {
				log.Printf("Receiving message...\n")
				msgType, msg, err := c.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
				// handling different message type
				switch msgType {
				case websocket.TextMessage:
					// TextMessage denotes a text data message. The text message payload is
					// interpreted as UTF-8 encoded text data.
					log.Printf("Text Message Received: %s\n", string(msg))

					// Send to broadcast channel to send messages to all client's send channel
					//c.hub.broadcast <- message
				case websocket.BinaryMessage:
					// BinaryMessage denotes a binary data message.
					log.Printf("Binary Message Received: Size: %d\n", len(msg))
					data := &message.DriverStatusPoll{}
					if err := proto.Unmarshal(msg, data); err != nil {
						log.Printf("Error decoding binary message: %v\n", err)
					} else {
						log.Printf("Decoded Message: (%d) Driver# %d currently at (%f, %f)\n", len(msg), data.DriverId, data.Lat, data.Lng)
					}

				case websocket.CloseMessage:
					// CloseMessage denotes a close control message. The optional message
					// payload contains a numeric code and text. Use the FormatCloseMessage
					// function to format a close message payload.
					log.Printf("Close Message Received: %v\n", string(msg))
					return
				case websocket.PingMessage:
					// PingMessage denotes a ping control message. The optional message payload
					// is UTF-8 encoded text.
					log.Printf("Ping Message Received: %v\n", string(msg))
				case websocket.PongMessage:
					// PongMessage denotes a pong control message. The optional message payload
					// is UTF-8 encoded text.
					log.Printf("Pong Message Received: %v\n", string(msg))
				default:
					log.Printf("Unknown Message Received\n")
				}
			}
		}(&s_wg)

		go func(s_wg *sync.WaitGroup) {

			//done := make(chan struct{})
			writeTicker := time.NewTicker(10 * time.Second)

			defer func() {
				log.Printf("Waitgroup done...\n")
				s_wg.Done()
				writeTicker.Stop()
			}()

			c.SetReadLimit(512)
			//c.SetReadDeadline(time.Now().Add(60 * time.Second))
			//c.SetPongHandler(func(string) error { c.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })

			writeTicker.Stop()
			for {

				select {
				case t := <-writeTicker.C:

					log.Printf("Sending message...%s\n", t.String())
					c.SetWriteDeadline(time.Now().Add(10 * time.Second))
					err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
					if err != nil {
						log.Println("write:", err)
					}
				case <-stop:
					log.Println("interrupt")

					// Cleanly close the connection by sending a close message and then
					// waiting (with timeout) for the server to close the connection.
					err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if err != nil {
						log.Println("write close:", err)
					}

					return
				}
			}
		}(&s_wg)

		// wait for all goroutines to finished
		s_wg.Wait()

		log.Printf("ending socket client...\n")
		c.Close()

	case "uniqueIDclient":
		//github.com/rs/xid:           bfvqibjq67r5lg4jir00
		common.GenXid()
		//github.com/segmentio/ksuid:  1Defmm7ODQ9BudqWXXwEPBkUxkA
		common.GenKsuid()
		//github.com/kjk/betterguid:   -LSTfCm53YQ4AvdYTvNW
		common.GenBetterGUID()
		//github.com/oklog/ulid:       01CXFASQ470KVRG0KBN1Y6504R
		common.GenUlid()

		//github.com/sony/sonyflake:   31e6a9a0e008901
		for i := 0; i < 5; i++ {
			common.GenSonyflake()
		}

		//github.com/chilts/sid:       1543481646241991400-2538134049391750617
		common.GenSid()

		//github.com/satori/go.uuid:   b3e26189-ba11-4335-9446-e80233330e2c
		common.GenUUIDv4()

	default:
		log.Printf("Unknown application mode: %s\n", *mode)
	}

	log.Printf("Exiting Main\n")
	os.Exit(0)
}
