package socket

import (
	"github.com/iknowhtml/locationtracker/pkg/common"
)

func NewWSRouter() []common.Route {

	wsRouter := []common.Route{
		//common.Route{"HandleWS", "GET", "/handle", WebSocketHandler},
		common.Route{"WSDriverStatus", "GET", "/driver/{id:[0-9]+}/status", WSDriverStatusHandler},
	}

	return wsRouter
}
