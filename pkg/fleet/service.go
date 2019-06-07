package fleet

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/iknowhtml/locationtracker/pkg/common"
)

type FleetService struct {
	fleetClient *FleetClient
}

func (fs *FleetService) Init() error {
	fs.fleetClient = new(FleetClient)
	err := fs.fleetClient.Init()
	if err != nil {
		return err
	}
	return nil
}

func (fs *FleetService) GetDriverFleetInfo(driverID int32) (*DriverFleetResponseObj, error) {
	respObj := DriverFleetResponseObj{}

	// check if driverID is zero
	if driverID == 0 {
		return nil, errors.New("Driver ID is zero.")
	}

	requestURI := common.Concate(fs.fleetClient.remoteAddr, API_STRING_DRIVERFLEET, "/", common.String(driverID))
	log.Printf("Request URI: %s \n", requestURI)
	res, body, _ := fs.fleetClient.request.Get(requestURI).End()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Error in sending request: " + requestURI)
	}

	err := json.Unmarshal([]byte(body), &respObj)
	if err != nil {
		return nil, err
	}
	log.Printf("Request Driver Fleet Info: %d, result: %v\n", driverID, respObj)

	if respObj.Data.DriverID == 0 || respObj.Data.ProviderID == 0 {
		return nil, errors.New("Driver Fleet info invalid or not found for id: " + common.String(driverID))
	}

	if respObj.Data.ActiveServiceTypeID == 0 || respObj.Data.ActiveServiceID == 0 {
		return nil, errors.New("No active service configured for id: " + common.String(driverID))
	}

	return &respObj, nil
}
