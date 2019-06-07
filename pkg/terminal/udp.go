package terminal

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iknowhtml/locationtracker/pkg/location"
	"github.com/iknowhtml/locationtracker/pkg/message"
)

// UDPServer holds the necessary structure for our
// UDP server.
type UDPServer struct {
	Addr   string
	Server *net.UDPConn
	Wg     *sync.WaitGroup
	Ch     chan message.DriverStatusPoll
}

func (u *UDPServer) New() *UDPServer {
	log.Printf("Initializing UDP server: %s\n", u.Addr)

	log.Printf("UDP Server initialized: %s\n", u.Addr)
	return u
}

/*
func (u *UDPServer) Init(addr string, wg *sync.WaitGroup, ch chan message.DriverStatusPoll) {
	log.Printf("Initializing UDP server: %s\n", addr)
	u.addr = addr
	u.wg = wg
	u.ch = ch
	log.Printf("UDP Server initialized: %s\n", addr)
}
*/

// Process will take the data from channel for processing.
func (u *UDPServer) Process() {
	log.Printf("Processing data: %s\n", u.Addr)
	for {
		processData(<-u.Ch)
	}
	log.Printf("Finished processing data: %s\n", u.Addr)
}

// Run starts the UDP server.
func (u *UDPServer) Run() {
	// signal the system to wait for server to finished running before exiting
	defer u.Wg.Done()

	log.Printf("Running server: %s\n", u.Addr)
	clientConns(u.Addr, u.Ch, u.Server)
	log.Printf("Server stopped: %s\n", u.Addr)
}

func clientConns(addr string, ch chan message.DriverStatusPoll, server *net.UDPConn) {
	log.Printf("%s - Handling client connections...\n", addr)
	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s - Resolved UDP address\n", addr)

	server, err = net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s - Listening to UDP port\n", addr)

	buf := make([]byte, 64)
	for {
		log.Printf("%s - Reading UDP packets\n", addr)
		n, c_addr, err := server.ReadFromUDP(buf)
		if err != nil {
			// if there is an error reading data from UDP, log it, and wait for next reading
			log.Printf("%s - Error encountered during reading: %s\n", addr, err.Error())
		} else {
			log.Printf("%s - Receiving data from: %s\n", addr, c_addr.String())
			handleConnections(n, buf, ch)
		}
	}
}

func handleConnections(n int, buf []byte, ch chan message.DriverStatusPoll) {
	log.Printf("Handling data: %d\n", n)

	data := &message.DriverStatusPoll{}
	err := proto.Unmarshal(buf[0:n], data)

	if err != nil {
		// if there is an decoding data into required DriverStatusPoll format, log it, and wait for next reading
		log.Printf("%d - Error encountered during decoding data: %s\n", n, err.Error())
	} else {
		log.Printf("%d - Received from Driver: %d at %d\n", n, data.DriverId, time.Now().Unix())
		//log.Printf("Fleet: %s\n", data.Fleet)
		//log.Printf("Driver Location: (%f, %f)\n", data.Lat, data.Lng)
		//log.Printf("Job Status: %d\n", data.Status)
		//log.Printf("Job ID: %d\n", data.JobId)
		//log.Println("---------------------------")

		// store received packet to channel
		ch <- *data
		log.Printf("%d - Stored data to channel\n", n)
	}

	log.Printf("Finished handling data: %d\n", n)
}

func processData(data message.DriverStatusPoll) {
	log.Printf("Processing data: %d at %d\n", data.DriverId, time.Now().Unix())
	locController := new(location.LocationController)
	locController.Init()

	/*
		// getting status based on data
		var d location.DriverStatus
		switch data.Status {
		// status is Available:
		case message.DriverStatusPoll_AVAILABLE:
			d = location.DriverStatus_AVAILABLE
		case message.DriverStatusPoll_NOTAVAILABLE:
			d = location.DriverStatus_NOTAVAILABLE
		case message.DriverStatusPoll_BUSY:
			d = location.DriverStatus_BUSY
		// else
		default:
			log.Panicf("Unknown driver status")
		}
	*/

	res, err := locController.UpdateDriverLocation(data.DriverId, data.Lat, data.Lng)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Successfully updated driver status: %d at %d - %v\n", data.DriverId, time.Now().Unix(), res)
	}

}

// Close ensures that the UDPServer is shut down gracefully.
func (u *UDPServer) Close() error {
	log.Printf("Closing server: %s\n", u.Addr)
	return u.Server.Close()
}
