package terminal

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iknowhtml/locationtracker/pkg/message"
)

type UDPClient struct {
	remoteAddr string
	client     *net.UDPConn
	wg         *sync.WaitGroup
}

func (c *UDPClient) Init(remoteAddr string, wg *sync.WaitGroup) {
	log.Printf("Initializing client...\n")
	c.remoteAddr = remoteAddr
	c.wg = wg
	log.Printf("Client initialized...\n")
}

// Run starts the UDP client.
func (c *UDPClient) Run(maxSendCount int) {
	log.Printf("Running client...\n")
	addr, err := net.ResolveUDPAddr("udp", c.remoteAddr)
	c.CheckError(err)

	if err == nil {
		c.connect(addr, maxSendCount)
	}
	log.Printf("Stopped client...\n")
}

func (c *UDPClient) connect(addr *net.UDPAddr, maxSendCount int) {
	log.Printf("Connecting client to server: %s\n", addr.String())
	conn, err := net.DialUDP("udp", nil, addr)
	c.CheckError(err)
	log.Printf("%s - Client connected to server: %s\n", conn.LocalAddr().String(), addr.String())
	c.client = conn
	defer c.Close()

	if maxSendCount < 1 {
		maxSendCount = 1
	}
	for i := 1; i <= maxSendCount; i++ {
		c.wg.Add(1)
		go func(id int32) {
			c.WriteToServer(id)
		}(int32(i))
	}
	c.wg.Wait()
	log.Printf("%s - Closing client connection to server: %s\n", c.client.LocalAddr().String(), addr.String())
}

func (c *UDPClient) WriteToServer(id int32) {
	log.Printf("%s - Writing data to server: %d\n", c.client.LocalAddr().String(), id)
	defer c.wg.Done()
	randJobID := rand.Intn(100)
	packet := CreatePacket(id, randJobID)
	data, err := proto.Marshal(packet)
	if err != nil {
		log.Panicf("marshalling error: ", err)
	}
	buf := []byte(data)

	_, err = c.client.Write(buf)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("%s - Finished writing data to server: %d\n", c.client.LocalAddr().String(), id)
	}
}

func CreatePacket(driverId int32, randJobID int) *message.DriverStatusPoll {
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

func (c *UDPClient) Close() error {
	return c.client.Close()
}

func (c *UDPClient) CheckError(err error) {
	if err != nil {
		log.Panicf("Error: ", err)
	}
}
