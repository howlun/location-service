package fleet

import (
	"log"

	"github.com/iknowhtml/locationtracker/pkg/config"
	"github.com/parnurzeal/gorequest"
)

type FleetClient struct {
	remoteAddr string
	request    *gorequest.SuperAgent
}

func (f *FleetClient) Init() error {
	log.Printf("Initializing Fleet Client...\n")

	// load system configuration based on environment, singleton pattern
	configuration, err := config.GetInstance("")
	if configuration == nil || err != nil {
		//log.Panicf("Failed to load configuration: %s\n", err.Error())
		return err
	}

	f.remoteAddr = configuration.Fleetserver.RemoteAddr
	f.request = gorequest.New()

	return nil
}
