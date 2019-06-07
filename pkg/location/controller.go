package location

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/iknowhtml/locationtracker/pkg/common"
	"github.com/iknowhtml/locationtracker/pkg/fleet"
)

type LocationController struct {
	locationService *LocationService
	fleetService    *fleet.FleetService
}

func (lc *LocationController) Init() error {
	lc.locationService = new(LocationService)
	err := lc.locationService.Init(nil)
	if err != nil {
		return err
	}
	lc.fleetService = new(fleet.FleetService)
	err = lc.fleetService.Init()
	if err != nil {
		return err
	}
	return nil
}

func (lc *LocationController) GetDriverStatus(driverID int32) (*GetObjectResponseObject, error) {

	res, err := lc.locationService.GetObject(Object_Collection_Fleet, driverID)
	if err != nil {
		return nil, err
	}

	// check if response object is empty
	if res == nil {
		return nil, errors.New("response object is empty")
	}

	return res, nil
}

// Update Driver Status for existing object which status is available or busy with extra fields
func (lc *LocationController) UpdateDriverStatus(
	driverID int32,
	cur_loc_lat float32,
	cur_loc_lng float32,
	available NearbySearch_Availability,
	jobID int32) (*SetObjectResponseObject, error) {

	// get current driver status
	driverExistObj, err := lc.locationService.GetObject(Object_Collection_Fleet, driverID)
	if err != nil {
		return nil, err
	}

	// check if response object is empty
	if driverExistObj == nil {
		return nil, errors.New("response object is empty")
	}

	var driverStatus DriverStatus
	switch available {
	case NearbySearch_Availability_2:
		driverStatus = DriverStatus_BUSY
	case NearbySearch_Availability_1:
		driverStatus = DriverStatus_AVAILABLE
		jobID = 0 //always set jobID to zero if set driver availability to Available
	default:
		driverStatus = DriverStatus_NOTAVAILABLE
	}

	// get fleet info
	driverFleetInfo, err := lc.fleetService.GetDriverFleetInfo(driverID)
	if err != nil {
		return nil, err
	}

	// Construct LocationObject in GeoJSON format
	locationObj := new(LocationObject)
	locationObj.Type = LocationObject_Type_Point
	locationObj.Coordinates = [2]float32{cur_loc_lat, cur_loc_lng}
	timeNow := time.Now().Unix()
	fields := LocationObject_Fields{}
	// if driver object exist
	if driverExistObj.Ok && driverExistObj.Fields.DriverID != 0 {

		// setup param object
		fields = LocationObject_Fields{
			"providerid":          driverFleetInfo.Data.ProviderID,
			"driverstatus":        int32(driverStatus),
			"jobid":               jobID,
			"activeserviceid":     driverFleetInfo.Data.ActiveServiceID,
			"activeservicetypeid": driverFleetInfo.Data.ActiveServiceTypeID,
			"priority":            driverFleetInfo.Data.Priority,
			"lastupdatedtime":     timeNow,
		}
	} else {

		fields = LocationObject_Fields{
			"driverid":            driverID,
			"providerid":          driverFleetInfo.Data.ProviderID,
			"driverstatus":        int32(driverStatus),
			"jobid":               jobID,
			"activeserviceid":     driverFleetInfo.Data.ActiveServiceID,
			"activeservicetypeid": driverFleetInfo.Data.ActiveServiceTypeID,
			"priority":            driverFleetInfo.Data.Priority,
			"lastupdatedtime":     timeNow,
		}
	}

	// Update object
	res, err := lc.locationService.SetObject(Object_Collection_Fleet, driverID, locationObj, fields)

	if err != nil {
		return nil, err
	}

	// check if ok is false
	if res.Ok == false {
		return nil, errors.New(res.Error)
	}

	log.Printf("Driver status updated: %v\n", res.Ok)
	return res, nil
}

// Set Driver Job Cancel
// if Driver object not found, return error "faield to retrieve object"
// if Driver status not Busy (not on job), return error "Set Driver job cancel not allowed, driver currently not on job"
// if Driver Job ID not match, return error "Set Driver job cancel not allowed, driver is on another job"

// Set Driver Job Complete
// if Driver object not found, return error "faield to retrieve object"
// if Driver status not Busy (not on job), return error "Set Driver job complete or cancel not allowed, driver currently not on job"
// if Driver Job ID not match, return error "Set Driver job complete or cancel not allowed, driver is on another job"
func (lc *LocationController) SetDriverJobCompleteOrCancel(
	driverID int32,
	jobID int32) (*SetFieldResponseObject, error) {

	// get current driver status
	driverExistObj, err := lc.locationService.GetObject(Object_Collection_Fleet, driverID)
	if err != nil {
		return nil, err
	}

	// check if response object is empty
	if driverExistObj == nil {
		return nil, errors.New("response object is empty")
	}

	// check if driver object found
	if driverExistObj.Ok != true {
		return nil, errors.New("failed to retrieve object")
	}

	// if Driver status not Busy (not on job), return error "Set Driver job complete not allowed, driver currently not on job"
	if driverExistObj.Ok && driverExistObj.Fields.Status != DriverStatus_BUSY {
		return nil, errors.New("Set Driver job complete or cancel not allowed, driver currently not on job")
	}

	// if Driver status not Busy (not on job), return error "Set Driver job complete not allowed, driver currently not on job"
	if driverExistObj.Ok && driverExistObj.Fields.Status == DriverStatus_BUSY && driverExistObj.Fields.JobID != jobID {
		return nil, errors.New("Set Driver job complete or cancel not allowed, driver is on another job")
	}

	timeNow := time.Now().Unix()
	fields := LocationObject_Fields{}

	// setup param object
	fields = LocationObject_Fields{
		"driverstatus":    int32(DriverStatus_AVAILABLE),
		"jobid":           0,
		"lastupdatedtime": timeNow,
	}

	// Update object
	res, err := lc.locationService.SetField(Object_Collection_Fleet, driverID, fields)
	if err != nil {
		return nil, err
	}

	// check if ok is false
	if res.Ok == false {
		return nil, errors.New(res.Error)
	}

	log.Printf("Driver Job Compelte or Cancel updated: %v\n", res.Ok)
	return res, nil
}

// Set Driver Busy and Job ID
// if Driver object not found, return error "faield to retrieve object"
// if Driver status is Busy, return error "Set Driver busy not allowed, driver currently on job"
// if Driver status is Not-Available, return error "Set Driver busy not allowed, driver currently not available"
func (lc *LocationController) SetAvailabilityBusy(
	driverID int32,
	jobID int32) (*SetFieldResponseObject, error) {

	// get current driver status
	driverExistObj, err := lc.locationService.GetObject(Object_Collection_Fleet, driverID)
	if err != nil {
		return nil, err
	}

	// check if response object is empty
	if driverExistObj == nil {
		return nil, errors.New("response object is empty")
	}

	// check if driver object found
	if driverExistObj.Ok != true {
		return nil, errors.New("failed to retrieve object")
	}

	// if Driver status is Busy, throw error "Set Driver busy not allowed, driver currently on job"
	// if Driver status is Not-Available, throw error "Set Driver busy not allowed, driver currently not available"
	if driverExistObj.Ok && driverExistObj.Fields.Status != DriverStatus_AVAILABLE {
		return nil, errors.New("Set Driver busy not allowed, driver currently on job or currently not available")
	}

	timeNow := time.Now().Unix()
	fields := LocationObject_Fields{}

	// setup param object
	fields = LocationObject_Fields{
		"driverstatus":    int32(DriverStatus_BUSY),
		"jobid":           jobID,
		"lastupdatedtime": timeNow,
	}

	// Update object
	res, err := lc.locationService.SetField(Object_Collection_Fleet, driverID, fields)
	if err != nil {
		return nil, err
	}

	// check if ok is false
	if res.Ok == false {
		return nil, errors.New(res.Error)
	}

	log.Printf("Driver availability updated: %v\n", res.Ok)
	return res, nil
}

// Set Driver availability with current location
// if Driver object not found, get Fleet Info to create Driver object
// else update availability and location
func (lc *LocationController) SetAvailability(
	driverID int32,
	cur_loc_lat float32,
	cur_loc_lng float32,
	available NearbySearch_Availability) (*SetObjectResponseObject, error) {

	// get current driver status
	driverExistObj, err := lc.locationService.GetObject(Object_Collection_Fleet, driverID)
	if err != nil {
		return nil, err
	}

	// check if response object is empty
	if driverExistObj == nil {
		return nil, errors.New("response object is empty")
	}

	var driverStatus DriverStatus
	switch available {
	case NearbySearch_Availability_1:
		driverStatus = DriverStatus_AVAILABLE
	default:
		driverStatus = DriverStatus_NOTAVAILABLE
	}

	// get fleet info
	driverFleetInfo, err := lc.fleetService.GetDriverFleetInfo(driverID)
	if err != nil {
		return nil, err
	}

	// Construct LocationObject in GeoJSON format
	locationObj := new(LocationObject)
	locationObj.Type = LocationObject_Type_Point
	locationObj.Coordinates = [2]float32{cur_loc_lat, cur_loc_lng}
	timeNow := time.Now().Unix()
	fields := LocationObject_Fields{}
	// if driver object exist
	if driverExistObj.Ok && driverExistObj.Fields.DriverID != 0 {
		// if driver is currently busy, throw error: cannot set availability when busy on job
		if driverExistObj.Fields.Status == DriverStatus_BUSY {
			return nil, errors.New("cannot change driver availability because driver is busy on job")
		}

		// setup param object
		fields = LocationObject_Fields{
			"providerid":          driverFleetInfo.Data.ProviderID,
			"driverstatus":        int32(driverStatus),
			"jobid":               0,
			"activeserviceid":     driverFleetInfo.Data.ActiveServiceID,
			"activeservicetypeid": driverFleetInfo.Data.ActiveServiceTypeID,
			"priority":            driverFleetInfo.Data.Priority,
			"lastupdatedtime":     timeNow,
		}
	} else {

		fields = LocationObject_Fields{
			"driverid":            driverID,
			"providerid":          driverFleetInfo.Data.ProviderID,
			"driverstatus":        int32(driverStatus),
			"jobid":               0,
			"activeserviceid":     driverFleetInfo.Data.ActiveServiceID,
			"activeservicetypeid": driverFleetInfo.Data.ActiveServiceTypeID,
			"priority":            driverFleetInfo.Data.Priority,
			"lastupdatedtime":     timeNow,
		}
	}

	// Update object
	res, err := lc.locationService.SetObject(Object_Collection_Fleet, driverID, locationObj, fields)

	if err != nil {
		return nil, err
	}

	// check if ok is false
	if res.Ok == false {
		return nil, errors.New(res.Error)
	}

	log.Printf("Driver availability updated: %v\n", res.Ok)
	return res, nil
}

// Update Driver Status for existing object which status is available or busy
func (lc *LocationController) UpdateDriverLocation(
	driverID int32,
	cur_loc_lat float32,
	cur_loc_lng float32) (interface{}, error) {

	// get current driver status
	driverExistObj, err := lc.locationService.GetObject(Object_Collection_Fleet, driverID)
	if err != nil {
		return nil, err
	}

	// check if response object is empty
	if driverExistObj == nil {
		return nil, errors.New("response object is empty")
	}

	// check if ok is false
	// example: id not found
	if driverExistObj.Ok == false {
		return nil, errors.New(driverExistObj.Error)
	}

	// if driver is currently not available, throw error: cannot update driver location when driver is not available
	if driverExistObj.Ok && driverExistObj.Fields.DriverID != 0 && driverExistObj.Fields.Status == DriverStatus_NOTAVAILABLE {
		return nil, errors.New("cannot update driver location when driver is not available")
	}

	// Construct LocationObject in GeoJSON format
	locationObj := new(LocationObject)
	locationObj.Type = LocationObject_Type_Point
	locationObj.Coordinates = [2]float32{cur_loc_lat, cur_loc_lng}
	timeNow := time.Now().Unix()
	fields := LocationObject_Fields{
		//"driverid":      driverID,
		//"providerid":      providerID,
		//"driverstatus":    int32(status),
		//"jobid":           jobId,
		//"activeserviceid": 0,
		//"activeservicetypeid": 0,
		//"priority": 0,
		"lastupdatedtime": timeNow,
	}

	// Update objects of fleet collection, with fleet type and driver id
	res, err := lc.locationService.SetObject(Object_Collection_Fleet, driverID, locationObj, fields)

	if err != nil {
		return nil, err
	}

	// check if ok is false
	if res.Ok == false {
		return nil, errors.New(res.Error)
	}

	log.Printf("Driver location updated: %v\n", res.Ok)
	return res, nil
}

func (lc *LocationController) SearchNearbyDriver(
	limit int32,
	from_lat float32,
	from_lng float32,
	search_tier int32,
	search_service_type_id int32,
	search_service_id int32,
	search_avail NearbySearch_Availability,
	search_priority NearbySearch_Priority) (*NearbyObjectMapObject, error) {

	// Where Conditions
	whereList := []WhereConditionFieldObject{}
	whereInList := []WhereInConditionFieldObject{}

	// Setup DriverStatus condition
	switch search_avail {
	case NearbySearch_Availability_1:
		var vals []interface{}
		vals = append(vals, int32(DriverStatus_AVAILABLE))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "driverstatus", Values: vals})
	case NearbySearch_Availability_0:
		var vals []interface{}
		vals = append(vals, int32(DriverStatus_NOTAVAILABLE))
		vals = append(vals, int32(DriverStatus_BUSY))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "driverstatus", Values: vals})
	default:
		// for case search all, do nothing
	}

	// Setup ActiveServiceTypeID condition
	if search_service_type_id != 0 {
		var vals []interface{}
		vals = append(vals, search_service_type_id)

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "activeservicetypeid", Values: vals})
	}

	// Setup ActiveServiceID condition
	if search_service_id != 0 {
		var vals []interface{}
		vals = append(vals, search_service_id)

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "activeserviceid", Values: vals})
	}

	// Setup Priority condition
	switch search_priority {
	case NearbySearch_Priority_1:
		var vals []interface{}
		vals = append(vals, int32(DriverPriority_YES))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "priority", Values: vals})
	case NearbySearch_Priority_0:
		var vals []interface{}
		vals = append(vals, int32(DriverPriority_NO))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "priority", Values: vals})
	default:
		// for case search all, do nothing
	}

	// Search nearby fleet objects from a point (lat, lng) with a radius
	nearbyObj, err := lc.locationService.NearbyObject(Object_Collection_Fleet, from_lat, from_lng, search_tier, limit, whereList, whereInList)

	if err != nil {
		return nil, err
	}
	/*
		// calculate distance in meters for each objects
		for i, o := range res.Objects {
			res.Objects[i].Distance = common.Distance(float64(from_lat), float64(from_lng), float64(o.Object.Coordinates[1]), float64(o.Object.Coordinates[0]))
			log.Printf("The distance of %s is %f\n", res.Objects[i].ID, res.Objects[i].Distance)
		}
	*/
	res := &NearbyObjectMapObject{}
	res = res.MapFrom(nearbyObj, from_lat, from_lng, nil)

	log.Printf("Search nearby: %v\n", res)
	return res, nil
}

func (lc *LocationController) SearchNearbyDriverByProviderId(
	limit int32,
	from_lat float32,
	from_lng float32,
	search_tier int32,
	filter_tier int32,
	providerID int32,
	search_service_type_id int32,
	search_service_id int32,
	search_avail NearbySearch_Availability,
	search_priority NearbySearch_Priority) (*NearbyObjectMapObject, error) {

	// Where Conditions
	whereList := []WhereConditionFieldObject{}
	whereInList := []WhereInConditionFieldObject{}

	// Setup ProviderID condition
	if providerID != 0 {
		var vals []interface{}
		vals = append(vals, providerID)

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "providerid", Values: vals})
	}

	// Setup DriverStatus condition
	switch search_avail {
	case NearbySearch_Availability_1:
		var vals []interface{}
		vals = append(vals, int32(DriverStatus_AVAILABLE))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "driverstatus", Values: vals})
	case NearbySearch_Availability_0:
		var vals []interface{}
		vals = append(vals, int32(DriverStatus_NOTAVAILABLE))
		vals = append(vals, int32(DriverStatus_BUSY))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "driverstatus", Values: vals})
	default:
		// for case search all, do nothing
	}

	// Setup ActiveServiceTypeID condition
	if search_service_type_id != 0 {
		var vals []interface{}
		vals = append(vals, search_service_type_id)

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "activeservicetypeid", Values: vals})
	}

	// Setup ActiveServiceID condition
	if search_service_id != 0 {
		var vals []interface{}
		vals = append(vals, search_service_id)

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "activeserviceid", Values: vals})
	}

	// Setup Priority condition
	switch search_priority {
	case NearbySearch_Priority_1:
		var vals []interface{}
		vals = append(vals, int32(DriverPriority_YES))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "priority", Values: vals})
	case NearbySearch_Priority_0:
		var vals []interface{}
		vals = append(vals, int32(DriverPriority_NO))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "priority", Values: vals})
	default:
		// for case search all, do nothing
	}
	filterObj := &NearbyObjectResponseObject{}
	if filter_tier > 0 {
		// Search nearby fleet objects from a point (lat, lng) with a radius
		var err error
		filterObj, err = lc.locationService.NearbyObject(Object_Collection_Fleet, from_lat, from_lng, filter_tier, limit, whereList, whereInList)
		if err != nil {
			return nil, err
		}
	}
	// Search nearby fleet objects from a point (lat, lng) with a radius
	nearbyObj, err := lc.locationService.NearbyObject(Object_Collection_Fleet, from_lat, from_lng, search_tier, limit, whereList, whereInList)
	if err != nil {
		return nil, err
	}
	/*
		// calculate distance in meters for each objects
		for i, o := range res.Objects {
			res.Objects[i].Distance = common.Distance(float64(from_lat), float64(from_lng), float64(o.Object.Coordinates[1]), float64(o.Object.Coordinates[0]))
			log.Printf("The distance of %s is %f\n", res.Objects[i].ID, res.Objects[i].Distance)
		}
	*/
	res := &NearbyObjectMapObject{}
	res = res.MapFrom(nearbyObj, from_lat, from_lng, filterObj)

	log.Printf("Search nearby: %v\n", res)
	return res, nil
}

func (lc *LocationController) DetectNearbyDriver(
	id int32,
	from_lat float32,
	from_lng float32,
	hookType Hook_Type,
	endPoints []string,
	fence_radius int32,
	search_service_type_id int32,
	search_service_id int32,
	search_avail NearbySearch_Availability,
	search_priority NearbySearch_Priority) (*HookFenceResponseObject, error) {

	// Where Conditions
	whereList := []WhereConditionFieldObject{}
	whereInList := []WhereInConditionFieldObject{}

	// Setup DriverStatus condition
	switch search_avail {
	case NearbySearch_Availability_1:
		var vals []interface{}
		vals = append(vals, int32(DriverStatus_AVAILABLE))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "driverstatus", Values: vals})
	case NearbySearch_Availability_0:
		var vals []interface{}
		vals = append(vals, int32(DriverStatus_NOTAVAILABLE))
		vals = append(vals, int32(DriverStatus_BUSY))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "driverstatus", Values: vals})
	default:
		// for case search all, do nothing
	}

	// Setup ActiveServiceTypeID condition
	if search_service_type_id != 0 {
		var vals []interface{}
		vals = append(vals, search_service_type_id)

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "activeservicetypeid", Values: vals})
	}

	// Setup ActiveServiceID condition
	if search_service_id != 0 {
		var vals []interface{}
		vals = append(vals, search_service_id)

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "activeserviceid", Values: vals})
	}

	// Setup Priority condition
	switch search_priority {
	case NearbySearch_Priority_1:
		var vals []interface{}
		vals = append(vals, int32(DriverPriority_YES))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "priority", Values: vals})
	case NearbySearch_Priority_0:
		var vals []interface{}
		vals = append(vals, int32(DriverPriority_NO))

		whereInList = append(whereInList, WhereInConditionFieldObject{FieldName: "priority", Values: vals})
	default:
		// for case search all, do nothing
	}

	// Detect List
	detectList := map[string]string{}                                                   // initialize detect list
	detectList[string(LocationDetect_Type_Inside)] = string(LocationDetect_Type_Inside) // add detect: inside
	detectList[string(LocationDetect_Type_Enter)] = string(LocationDetect_Type_Enter)   // add detect: enter

	// Command List
	commandList := map[string]string{} // initialize command list

	// Detect nearby fleet objects from a point (lat, lng) with a radius
	res, err := lc.locationService.SetHookSearchFence(
		endPoints,
		common.Concate(HookPrefix, strings.Title(string(hookType))), string(LocationSearch_Type_Nearby),
		Object_Collection_Fleet, id, from_lat, from_lng, fence_radius, detectList, commandList, whereList, whereInList)

	if err != nil {
		return nil, err
	}

	log.Printf("Start detect nearby: %v\n", res)
	return res, nil
}

func (lc *LocationController) StopDetectNearbyDriver(
	id int32,
	hookType Hook_Type) (*HookFenceResponseObject, error) {

	// Stop Detect nearby fleet objects
	res, err := lc.locationService.DelHookSearchFence(common.Concate(HookPrefix, strings.Title(string(hookType))), string(LocationSearch_Type_Nearby), Object_Collection_Fleet, id)

	if err != nil {
		return nil, err
	}

	log.Printf("Stop detect nearby: %v\n", res)
	return res, nil
}
