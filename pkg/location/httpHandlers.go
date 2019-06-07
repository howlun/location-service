package location

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/iknowhtml/locationtracker/pkg/common"
	"github.com/iknowhtml/locationtracker/pkg/config"
)

// driver id (id)
// POST body: { "avail": [1|0], "lat": 50.1000, "lng": 101.1000}
func HandleSetDriverAvailability(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		common.HandleMethodNotAllowedResponse(w, "")
		return
	}

	vars := mux.Vars(r)

	var driverID int
	if vars["id"] != "" && vars["id"] != "0" {
		driverID, _ = strconv.Atoi(vars["id"])
	} else {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Driver ID is missing, but required")
		return
	}

	// reading POST body
	log.Println("Decoding request json body")
	decoder := json.NewDecoder(r.Body)
	var reqObj DriverAvailabilityRequestObject
	err := decoder.Decode(&reqObj)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}
	log.Println(reqObj)

	if reqObj.Lat == 0 {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Driver Location (Latitude) is missing, but required")
		return
	}

	if reqObj.Lng == 0 {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Driver Location (Longitude) is missing, but required")
		return
	}

	var available NearbySearch_Availability
	switch reqObj.Availability {
	case "1":
		available = NearbySearch_Availability_1
	case "0":
		available = NearbySearch_Availability_0
	default:
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Driver Availability is missing, but required")
		return
	}

	locController := new(LocationController)
	err = locController.Init()
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	res, err := locController.SetAvailability(int32(driverID), reqObj.Lat, reqObj.Lng, available)
	if err != nil {
		// send a internal server error back to the caller'
		common.HandleServerErrorResponse(w, err)
		return
	}

	if res != nil && res.Ok {
		common.HandleStatusOKResponse(w, &common.EmptyResultObject{})
	} else {
		common.HandleStatus400Response(w, res.Error)
	}
}

// query param: driver id (id) (required)
func HandleGetDriverStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		common.HandleMethodNotAllowedResponse(w, "")
		return
	}

	vars := mux.Vars(r)

	var driverID int
	if vars["id"] != "" && vars["id"] != "0" {
		driverID, _ = strconv.Atoi(vars["id"])
	} else {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Driver ID is missing, but required")
		return
	}

	locController := new(LocationController)
	err := locController.Init()
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	res, err := locController.GetDriverStatus(int32(driverID))
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	if res != nil && res.Ok {
		common.HandleStatusOKResponse(w, &DriverStatusObject{DriverStatus: res})
	} else {
		common.HandleStatus400Response(w, res.Error)
	}
}

// url: driver id ("id") (required)
// post: lat (lat) (required)
// post: lng (lng) (required)
// post: availability (avail = 1|0) (required)
// post: job id ("jobid") (required) -- if avail = 1, jobid is always zero
func HandleSetDriverStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		common.HandleMethodNotAllowedResponse(w, "")
		return
	}

	vars := mux.Vars(r)

	var driverID int
	if vars["id"] != "" && vars["id"] != "0" {
		driverID, _ = strconv.Atoi(vars["id"])
	} else {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Driver ID is missing, but required")
		return
	}

	// reading POST body
	log.Println("Decoding request body")
	decoder := json.NewDecoder(r.Body)
	var reqObj SetDriverStatusRequestObject
	err := decoder.Decode(&reqObj)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}
	log.Println(reqObj)

	if reqObj.Lat == 0 {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Location (Latitude) is missing, but required")
		return
	}

	if reqObj.Lng == 0 {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Location (Longitude) is missing, but required")
		return
	}

	if reqObj.JobId == 0 && reqObj.Availability != "1" {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Job ID is missing, but required")
		return
	}

	var searchAvail NearbySearch_Availability
	switch reqObj.Availability {
	case "2":
		searchAvail = NearbySearch_Availability_2
	case "1":
		searchAvail = NearbySearch_Availability_1
	case "0":
		searchAvail = NearbySearch_Availability_0
	default:
		searchAvail = NearbySearch_Availability_All
	}

	locController := new(LocationController)
	err = locController.Init()
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	res, err := locController.UpdateDriverStatus(int32(driverID), reqObj.Lat, reqObj.Lng, searchAvail, reqObj.JobId)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	if res != nil && res.Ok {
		common.HandleStatusOKResponse(w, &common.EmptyResultObject{})
	} else {
		common.HandleStatus400Response(w, res.Error)
	}
}

// tier (tier = 1|2|3) (required)
// Url Param: lat (e_lat) (required)
// Url Param: lng (e_lng) (required)
// providerid (provider) (optional)
// availability (avail = 1|0) (optional)
// service type id (srvtype = 0) (optional)
// service id (srv = 0) (optional)
// priority (priority = 1|0) (optional)
func HandleGetNearby(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		common.HandleMethodNotAllowedResponse(w, "")
		return
	}

	// get Url Param
	queryValues := r.URL.Query()
	log.Println(queryValues)

	var searchTier int32
	if queryValues.Get("tier") == "" {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Scan Tier is missing, but required")
		return
	}

	// load system configuration based on environment, singleton pattern
	configuration, err := config.GetInstance("")
	if configuration == nil {
		log.Panicf("Failed to load configuration: %s\n", err.Error())
	}
	var filterTier int32
	switch queryValues.Get("tier") {
	case "1":
		filterTier = 0 // default 0 to turn off filter
		searchTier = configuration.Locationremoteserver.SearchTier1Meter
	case "2":
		filterTier = configuration.Locationremoteserver.SearchTier1Meter
		searchTier = configuration.Locationremoteserver.SearchTier2Meter
	case "3":
		filterTier = configuration.Locationremoteserver.SearchTier2Meter
		searchTier = configuration.Locationremoteserver.SearchTier3Meter
	default:
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Scan Tier is missing or invalid")
		return
	}

	if searchTier == 0 {
		common.HandleStatus400Response(w, "Scan Tier is set but no range is defined")
		return
	}

	if queryValues.Get("e_lat") == "" {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Latitude is missing, but required")
		return
	}
	if queryValues.Get("e_lng") == "" {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Longitude is missing, but required")
		return
	}
	e_lat, err := strconv.ParseFloat(queryValues.Get("e_lat"), 32)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, err.Error())
		return
	}
	e_lng, err := strconv.ParseFloat(queryValues.Get("e_lng"), 32)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, err.Error())
		return
	}
	var searchAvail NearbySearch_Availability
	switch queryValues.Get("avail") {
	case "1":
		searchAvail = NearbySearch_Availability_1
	case "0":
		searchAvail = NearbySearch_Availability_0
	default:
		searchAvail = NearbySearch_Availability_All
	}

	var serviceTypeID int64
	if queryValues.Get("srvtype") != "" {
		serviceTypeID, err = strconv.ParseInt(queryValues.Get("srvtype"), 10, 32)
		if err != nil {
			// send a internal server error back to the caller
			common.HandleStatus400Response(w, err.Error())
			return
		}
	}

	var serviceID int64
	if queryValues.Get("srv") != "" {
		serviceID, err = strconv.ParseInt(queryValues.Get("srv"), 10, 32)
		if err != nil {
			// send a internal server error back to the caller
			common.HandleStatus400Response(w, err.Error())
			return
		}
	}

	var searchPriority NearbySearch_Priority
	switch queryValues.Get("priority") {
	case "1":
		searchPriority = NearbySearch_Priority_1
	case "0":
		searchPriority = NearbySearch_Priority_0
	default:
		searchPriority = NearbySearch_Priority_All
	}

	var providerid int64
	if queryValues.Get("provider") != "" {
		providerid, err = strconv.ParseInt(queryValues.Get("provider"), 10, 32)
		if err != nil {
			// send a internal server error back to the caller
			common.HandleStatus400Response(w, err.Error())
			return
		}
	}

	locController := new(LocationController)
	err = locController.Init()
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	res, err := locController.SearchNearbyDriverByProviderId(Search_Limit, float32(e_lat), float32(e_lng), searchTier, filterTier, int32(providerid), int32(serviceTypeID), int32(serviceID), searchAvail, searchPriority)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	if res != nil && res.Ok {
		common.HandleStatusOKResponse(w, &NearbyDriversObject{NearbyDrivers: res})
	} else {
		common.HandleStatus400Response(w, res.Error)
	}
}

// post: lat (e_lat) (required)
// post: lng (e_lng) (required)
// post: driver id ("id") (optional)
// post: availability (avail = 1|0) (optional)
// post: service type id (srvtype = 0) (optional)
// post: service id (srv = 0) (optional)
// post: priority (priority = 1|0) (optional)
func HandleStartDetectArriving(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		common.HandleMethodNotAllowedResponse(w, "")
		return
	}
	//vars := mux.Vars(r)

	var fenceRadius int32
	var endPoints []string
	// load system configuration based on environment, singleton pattern
	configuration, err := config.GetInstance("")
	if configuration == nil {
		log.Panicf("Failed to load configuration: %s\n", err.Error())
	}
	fenceRadius = configuration.Locationremoteserver.DetectArrivingMeter
	endPoints = configuration.Locationremoteserver.HookEndpoints

	// reading POST body
	log.Println("Decoding request body")
	decoder := json.NewDecoder(r.Body)
	var reqObj StartNearbyFenceRequestObject
	err = decoder.Decode(&reqObj)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}
	log.Println(reqObj)

	if reqObj.E_lat == 0 {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Location (Latitude) is missing, but required")
		return
	}

	if reqObj.E_lng == 0 {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Location (Longitude) is missing, but required")
		return
	}

	var searchAvail NearbySearch_Availability
	switch reqObj.Availability {
	case "1":
		searchAvail = NearbySearch_Availability_1
	case "0":
		searchAvail = NearbySearch_Availability_0
	default:
		searchAvail = NearbySearch_Availability_All
	}

	var searchPriority NearbySearch_Priority
	switch reqObj.Priority {
	case "1":
		searchPriority = NearbySearch_Priority_1
	case "0":
		searchPriority = NearbySearch_Priority_0
	default:
		searchPriority = NearbySearch_Priority_All
	}

	locController := new(LocationController)
	err = locController.Init()
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	res, err := locController.DetectNearbyDriver(
		reqObj.ID, reqObj.E_lat, reqObj.E_lng, Hook_Type_Arriving, endPoints, fenceRadius, reqObj.SearchServiceTypeID, reqObj.SearchServiceID, searchAvail, searchPriority)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	if res != nil && res.Ok {
		common.HandleStatusOKResponse(w, &NearbyDriversObject{NearbyDrivers: res})
	} else {
		common.HandleStatus400Response(w, res.Error)
	}
}

// driver id (id) (optional)
func HandleStopDetectArriving(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		common.HandleMethodNotAllowedResponse(w, "")
		return
	}

	var err error

	// reading POST body
	log.Println("Decoding request body")
	decoder := json.NewDecoder(r.Body)
	var reqObj StopNearbyFenceRequestObject
	err = decoder.Decode(&reqObj)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}
	log.Println(reqObj)

	locController := new(LocationController)
	err = locController.Init()
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	res, err := locController.StopDetectNearbyDriver(reqObj.ID, Hook_Type_Arriving)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	if res != nil && res.Ok {
		common.HandleStatusOKResponse(w, &NearbyDriversObject{NearbyDrivers: res})
	} else {
		common.HandleStatus400Response(w, res.Error)
	}
}

// post: lat (e_lat) (required)
// post: lng (e_lng) (required)
// post: driver id (id) (optional)
// post: availability (avail = 1|0) (optional)
// post: service type id (srvtype = 0) (optional)
// post: service id (srv = 0) (optional)
// post: priority (priority = 1|0) (optional)
func HandleStartDetectArrived(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		common.HandleMethodNotAllowedResponse(w, "")
		return
	}
	//vars := mux.Vars(r)

	var fenceRadius int32
	var endPoints []string
	// load system configuration based on environment, singleton pattern
	configuration, err := config.GetInstance("")
	if configuration == nil {
		log.Panicf("Failed to load configuration: %s\n", err.Error())
	}
	fenceRadius = configuration.Locationremoteserver.DetectArrivedMeter
	endPoints = configuration.Locationremoteserver.HookEndpoints

	// reading POST body
	log.Println("Decoding request body")
	decoder := json.NewDecoder(r.Body)
	var reqObj StartNearbyFenceRequestObject
	err = decoder.Decode(&reqObj)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}
	log.Println(reqObj)

	if reqObj.E_lat == 0 {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Location (Latitude) is missing, but required")
		return
	}

	if reqObj.E_lng == 0 {
		// send a internal server error back to the caller
		common.HandleStatus400Response(w, "Location (Longitude) is missing, but required")
		return
	}

	var searchAvail NearbySearch_Availability
	switch reqObj.Availability {
	case "1":
		searchAvail = NearbySearch_Availability_1
	case "0":
		searchAvail = NearbySearch_Availability_0
	default:
		searchAvail = NearbySearch_Availability_All
	}

	var searchPriority NearbySearch_Priority
	switch reqObj.Priority {
	case "1":
		searchPriority = NearbySearch_Priority_1
	case "0":
		searchPriority = NearbySearch_Priority_0
	default:
		searchPriority = NearbySearch_Priority_All
	}

	locController := new(LocationController)
	err = locController.Init()
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	res, err := locController.DetectNearbyDriver(
		reqObj.ID, reqObj.E_lat, reqObj.E_lng, Hook_Type_Arrived, endPoints, fenceRadius, reqObj.SearchServiceTypeID, reqObj.SearchServiceID, searchAvail, searchPriority)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	if res != nil && res.Ok {
		common.HandleStatusOKResponse(w, &NearbyDriversObject{NearbyDrivers: res})
	} else {
		common.HandleStatus400Response(w, res.Error)
	}
}

// driver id (id) (optional)
func HandleStopDetectArrived(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		common.HandleMethodNotAllowedResponse(w, "")
		return
	}
	var err error

	// reading POST body
	log.Println("Decoding request body")
	decoder := json.NewDecoder(r.Body)
	var reqObj StopNearbyFenceRequestObject
	err = decoder.Decode(&reqObj)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}
	log.Println(reqObj)

	locController := new(LocationController)
	err = locController.Init()
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	res, err := locController.StopDetectNearbyDriver(reqObj.ID, Hook_Type_Arrived)
	if err != nil {
		// send a internal server error back to the caller
		common.HandleServerErrorResponse(w, err)
		return
	}

	if res != nil && res.Ok {
		common.HandleStatusOKResponse(w, &NearbyDriversObject{NearbyDrivers: res})
	} else {
		common.HandleStatus400Response(w, res.Error)
	}
}
