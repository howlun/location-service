package location

import (
	"log"

	"github.com/iknowhtml/locationtracker/pkg/common"
)

type LocationObject_Fields map[string]interface{}

type LocationObject struct {
	Type        LocationObject_Type `json:"type"`
	Coordinates [2]float32          `json:"coordinates"`
}

type LocationObject_Properties struct {
	DriverID             int32        `json:"driverid"`
	ProviderID           int32        `json:"providerid"`
	Status               DriverStatus `json:"driverstatus"`
	JobID                int32        `json:"jobid"`
	ActiveServiceID      int32        `json:"activeserviceid"`
	ActiveServiceTypeID  int32        `json:"activeservicetypeid"`
	Priority             int32        `json:"priority"`
	LastUpdatedTimestamp int64        `json:"lastupdatedtime"`
}

type LocationResponseObject struct {
	Type        LocationObject_Type `json:"type"`
	Coordinates [2]float32          `json:"coordinates"`
}

type GetObjectResponseObject struct {
	Ok               bool                      `json:"ok"`
	ObjectCollection Object_Collection         `json:"collection,omitempty"`
	Object           LocationResponseObject    `json:"object,omitempty"`
	Fields           LocationObject_Properties `json:"fields,omitempty"`
	Error            string                    `json:"err,omitempty"`
	Elapsed          string                    `json:"elapsed"`
}

type DriverStatusObject struct {
	DriverStatus interface{} `json:"driverstatus"`
}

func (o *DriverStatusObject) SetResult(result interface{}) {
	o.DriverStatus = result
}

/*
type DriverStatusRequestObject struct {
	Fleet     string       `json:"fleet,omitempty"`
	DriverId  int32        `json:"driverId,omitempty"`
	Lat       float32      `json:"lat,omitempty"`
	Lng       float32      `json:"lng,omitempty"`
	Status    DriverStatus `json:"status,omitempty"`
	JobId     int64        `json:"jobId,omitempty"`
	Timestamp int64        `json:"timestamp,omitempty"`
}
*/
type DriverAvailabilityRequestObject struct {
	Availability string  `json:"avail,omitempty"`
	Lat          float32 `json:"lat,omitempty"`
	Lng          float32 `json:"lng,omitempty"`
}

type SetFieldResponseObject struct {
	Ok      bool   `json:"ok"`
	Error   string `json:"err,omitempty"`
	Elapsed string `json:"elapsed"`
}

type SetObjectResponseObject struct {
	Ok      bool   `json:"ok"`
	Error   string `json:"err,omitempty"`
	Elapsed string `json:"elapsed"`
}

type ObjectsResponseObject struct {
	ID       string                 `json:"id,omitempty"`
	Object   LocationResponseObject `json:"object,omitempty"`
	Fields   []interface{}          `json:"fields,omitempty"`
	Distance float64                `json:"distance,omitempty"` // in meters
}

type NearbyObjectResponseObject struct {
	Ok               bool                    `json:"ok"`
	ObjectCollection Object_Collection       `json:"collection,omitempty"`
	Fields           []string                `json:"fields,omitempty"`
	Objects          []ObjectsResponseObject `json:"objects,omitempty"`
	Error            string                  `json:"err,omitempty"`
	Count            int32                   `json:"count"`
	Cursor           int32                   `json:"cursor,omitempty"`
	Elapsed          string                  `json:"elapsed"`
}

func (o *NearbyObjectResponseObject) RemoveObject(objID string) {
	if objID != "" && len(o.Objects) > 0 {
		foundIndex := -1
		for i, o := range o.Objects {
			if o.ID == objID {
				foundIndex = i
			}
		}

		if foundIndex != -1 {
			o.Objects[foundIndex] = o.Objects[0] // copy first element to the i index
			o.Objects = o.Objects[1:]            // return a copy of array starting from 2 item in the array
		}
	}
}

type ObjectsMapObject struct {
	ID       string                    `json:"id,omitempty"`
	Object   LocationResponseObject    `json:"object,omitempty"`
	Fields   LocationObject_Properties `json:"fields,omitempty"`
	Distance float64                   `json:"distance,omitempty"` // in meters
}

type NearbyObjectMapObject struct {
	Ok               bool               `json:"ok"`
	ObjectCollection Object_Collection  `json:"collection,omitempty"`
	Objects          []ObjectsMapObject `json:"objects,omitempty"`
	Error            string             `json:"err,omitempty"`
	Count            int32              `json:"count"`
	Cursor           int32              `json:"cursor,omitempty"`
	Elapsed          string             `json:"elapsed"`
}

func (o *NearbyObjectMapObject) MapFrom(from *NearbyObjectResponseObject, from_lat float32, from_lng float32, filter *NearbyObjectResponseObject) *NearbyObjectMapObject {

	newObj := &NearbyObjectMapObject{
		Ok:               from.Ok,
		ObjectCollection: from.ObjectCollection,
		Error:            from.Error,
		Count:            0,
		Cursor:           from.Cursor,
		Elapsed:          from.Elapsed}

	if from != nil && from.Count > 0 && len(from.Objects) > 0 {
		log.Printf("NearbyObjectMapObject.MapFrom: Objects: %v\n", from.Objects)

		// filter the Filter Objects
		if filter != nil && filter.Count > 0 && len(filter.Objects) > 0 {
			log.Printf("NearbyObjectMapObject.MapFrom: Filter Objects: %v\n", filter.Objects)
			// loop through each fitler objects
			for _, f := range filter.Objects {
				from.RemoveObject(f.ID)
			}
		}

		log.Printf("NearbyObjectMapObject.MapFrom: Objects after filter: %v\n", from.Objects)

		// Construct new Map Object with possibly new array count after filter
		newCount := len(from.Objects)
		newObj.Count = int32(newCount)
		objs := make([]ObjectsMapObject, newCount)
		for i, fo := range from.Objects {
			log.Printf("NearbyObjectMapObject.MapFrom: Each object: %v\n", fo)
			to := ObjectsMapObject{
				ID:       fo.ID,
				Object:   fo.Object,
				Fields:   LocationObject_Properties{},
				Distance: common.Distance(float64(from_lat), float64(from_lng), float64(fo.Object.Coordinates[1]), float64(fo.Object.Coordinates[0]))}

			// loop through all fields
			for j, fname := range from.Fields {
				switch fname {
				case "providerid":
					to.Fields.ProviderID = int32(fo.Fields[j].(float64))
				case "driverstatus":
					to.Fields.Status = DriverStatus(int(fo.Fields[j].(float64)))
				case "jobid":
					to.Fields.JobID = int32(fo.Fields[j].(float64))
				case "activeserviceid":
					to.Fields.ActiveServiceID = int32(fo.Fields[j].(float64))
				case "activeservicetypeid":
					to.Fields.ActiveServiceTypeID = int32(fo.Fields[j].(float64))
				case "priority":
					to.Fields.Priority = int32(fo.Fields[j].(float64))
				case "lastupdatedtime":
					to.Fields.LastUpdatedTimestamp = int64(fo.Fields[j].(float64))
				case "driverid":
					to.Fields.DriverID = int32(fo.Fields[j].(float64))
				default:
					log.Panicf("ObjectsMapObject: Unknow field for mapping...")
				}
			}

			objs[i] = to
		}
		newObj.Objects = objs
	}

	return newObj
}

type NearbyDriversObject struct {
	NearbyDrivers interface{} `json:"nearbydrivers"`
}

func (o *NearbyDriversObject) SetResult(result interface{}) {
	o.NearbyDrivers = result
}

type WhereConditionFieldObject struct {
	FieldName string
	Min       interface{}
	Max       interface{}
}

type WhereInConditionFieldObject struct {
	FieldName string
	Values    []interface{}
}

type StartNearbyFenceRequestObject struct {
	E_lat               float32 `json:"e_lat,omitempty"`
	E_lng               float32 `json:"e_lng,omitempty"`
	ID                  int32   `json:"id,omitempty"`
	Availability        string  `json:"avail,omitempty"`
	SearchServiceID     int32   `json:"srv"`
	SearchServiceTypeID int32   `json:"srvtype"`
	Priority            string  `json:"priority"`
}

type SetDriverStatusRequestObject struct {
	Availability string  `json:"avail,omitempty"`
	Lat          float32 `json:"lat,omitempty"`
	Lng          float32 `json:"lng,omitempty"`
	JobId        int32   `json:"jobid,omitempty"`
}

type StopNearbyFenceRequestObject struct {
	ID int32 `json:"id,omitempty"`
}

type HookFenceResponseObject struct {
	Ok      bool   `json:"ok"`
	Error   string `json:"err,omitempty"`
	Elapsed string `json:"elapsed"`
}

var DetectList map[string]string
