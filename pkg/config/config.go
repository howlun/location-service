package config

import (
	"log"
	"sync"

	"github.com/iknowhtml/locationtracker/pkg/common"
	"github.com/spf13/viper"
)

type MonitorServerConfig struct {
	Addr string `json:"addr"`
}

type CORSConfig struct {
	AllowedOrigins     []string `json:"allowedorigins"`
	AllowedMethods     []string `json:"allowedmethods"`
	AllowedHeaders     []string `json:"allowedheaders"`
	AllowCredentials   bool     `json:"allowcredentials"`
	Debug              bool     `json:"debug"`
	OptionsPassthrough bool     `json:"optionspassthrough"`
	MaxAge             int32    `json:"maxage"`
}

type UDPServerConfig struct {
	Addr string `json:"addr"`
}

type HTTPServerConfig struct {
	Addr string `json:"addr"`
}

type LocationRemoteServerConfig struct {
	RemoteAddr          string   `json:"remoteaddr"`
	HookEndpoints       []string `json:"hookendpoints"`
	SearchTier1Meter    int32    `json:"searchtier1meter"`
	SearchTier2Meter    int32    `json:"searchtier2meter"`
	SearchTier3Meter    int32    `json:"searchtier3meter"`
	DetectArrivingMeter int32    `json:"detectarrivingmeter"`
	DetectArrivedMeter  int32    `json:"detectarrivedmeter"`
}

type AuthServerConfig struct {
	RemoteAddr string `json:"remoteaddr"`
}

type FleetServerConfig struct {
	RemoteAddr string `json:"remoteaddr"`
}

type SocketServerConfig struct {
	Addr string `json:"addr"`
}

type Configuration struct {
	Corsconfig           CORSConfig                 `json:"corsconfig"`
	Udpserver            UDPServerConfig            `json:"udpserver"`
	Httpserver           HTTPServerConfig           `json:"httpserver"`
	Locationremoteserver LocationRemoteServerConfig `json:"locationremoteserver"`
	Authserver           AuthServerConfig           `json:"authserver"`
	Fleetserver          FleetServerConfig          `json:"fleetserver"`
	Socketserver         SocketServerConfig         `json:"socketserver"`
}

var c *Configuration
var once sync.Once

func GetInstance(env string) (*Configuration, error) {
	log.Printf("Retrieving configuration...\n")

	var err error
	once.Do(func() {
		log.Printf("Creating new configuration instance...\n")

		c = &Configuration{}

		// initialize the viper variables
		v := viper.New()
		v.SetConfigType("yaml")
		//v.AddConfigPath("/var/lib/locationtracker/")
		//v.AddConfigPath("$GOPATH/src/github.com/iknowhtml/locationtracker/")
		v.AddConfigPath(".")

		switch env {
		case string(common.EnvType_Prod):
			v.SetConfigName(string(common.EnvType_Prod))
		default:
			v.SetConfigName(string(common.EnvType_Dev))
		}

		// Set defaults
		v.SetDefault("udpserver.addr", ":9000")
		v.SetDefault("httpserver.addr", ":8000")
		v.SetDefault("socketserver.addr", ":8010")
		v.SetDefault("locationremoteserver.remoteaddr", "35.185.186.230:9851")
		v.SetDefault("hookendpoints", []string{"http://localhost:8000/ep1", "http://localhost:8000/ep2"})
		v.SetDefault("locationremoteserver.searchtier1meter", 5000)
		v.SetDefault("locationremoteserver.searchtier2meter", 10000)
		v.SetDefault("locationremoteserver.searchtier3meter", 0)
		v.SetDefault("locationremoteserver.detectarrivingmeter", 500)
		v.SetDefault("locationremoteserver.detectarrivedmeter", 50)
		v.SetDefault("authserver.remoteaddr", "35.240.167.230:8080")

		// Read configuration
		log.Printf("Reading configuration for %s env...\n", env)
		err = v.ReadInConfig()

		if err == nil {
			err = v.Unmarshal(&c)
			log.Printf("Configuration file (%s env) has been loaded: %v\n", env, c)
		}
	})

	if err != nil {
		return nil, err
	}

	return c, nil
}
