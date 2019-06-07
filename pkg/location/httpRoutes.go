package location

import (
	"github.com/iknowhtml/locationtracker/pkg/common"
)

func NewRouter() []common.Route {

	fleetRouter := []common.Route{
		common.Route{"SetDriverAvailability", "POST", "/driver/{id:[0-9]+}/availability", HandleSetDriverAvailability},
		common.Route{"GetDriverStatus", "GET", "/driver/{id:[0-9]+}/status", HandleGetDriverStatus},
		common.Route{"GetNearby", "GET", "/driver/nearby", HandleGetNearby},
		common.Route{"SetDetectArriving", "POST", "/driver/startdetectarriving", HandleStartDetectArriving},
		common.Route{"DelDetectArriving", "DELETE", "/driver/stopdetectarriving", HandleStopDetectArriving},
		common.Route{"SetDetectArrived", "POST", "/driver/startdetectarrived", HandleStartDetectArrived},
		common.Route{"DelDetectArrived", "DELETE", "/driver/stopdetectarrived", HandleStopDetectArrived},
		common.Route{"SetDriverStatus", "POST", "/driver/{id:[0-9]+}/status", HandleSetDriverStatus},
	}

	return fleetRouter
}
