package caching

import (
	"log"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

var p *redis.Pool
var once sync.Once

func NewPool(server, password string, outputJSON bool) *redis.Pool {
	log.Printf("Retriving connection pool...\n")

	once.Do(func() {
		log.Printf("Creating new configuration instance...\n")

		p = &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", server)
				if err != nil {
					return nil, err
				}

				// set auth
				if password != "" {
					if _, err := c.Do("AUTH", password); err != nil {
						c.Close()
						return nil, err
					}
				}

				// set server to return JSON
				if outputJSON {
					if _, err := c.Do("OUTPUT", "json"); err != nil {
						//log.Panicf("Error setting output to JSON: %s\n", err)
						return nil, err
					}
				}

				log.Printf("Successfully connected to location server: %s\n", server)
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}
	})

	return p
}
