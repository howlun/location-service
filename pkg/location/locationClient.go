package location

import (
	"log"

	"github.com/gomodule/redigo/redis"
	"github.com/iknowhtml/locationtracker/pkg/caching"
	"github.com/iknowhtml/locationtracker/pkg/config"
)

type LocationClient struct {
	remoteAddr string
	pool       *redis.Pool
}

func (l *LocationClient) Init() error {
	log.Printf("Initializing Location Client...\n")

	// load system configuration based on environment, singleton pattern
	configuration, err := config.GetInstance("")
	if configuration == nil || err != nil {
		//log.Panicf("Failed to load configuration: %s\n", err.Error())
		return err
	}

	l.remoteAddr = configuration.Locationremoteserver.RemoteAddr
	l.pool = caching.NewPool(l.remoteAddr, "", true)

	return nil
}
