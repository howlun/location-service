package location

const (
	PathPrefix string = "fleet/"
	HookPrefix string = "HKDRIVER"
)

// HookType
type Hook_Type string

const (
	Hook_Type_Arriving Hook_Type = "ARRIVING"
	Hook_Type_Arrived  Hook_Type = "ARRIVED"
)

// SearchLimit
const (
	Search_Limit int32 = 20
)

// ObjectCollection
type Object_Collection string

const (
	Object_Collection_Fleet Object_Collection = "fleet"
	Object_Collection_POI   Object_Collection = "poi"
)

// LocationSearch_Type
type LocationSearch_Type string

const (
	LocationSearch_Type_Nearby     LocationDetect_Type = "nearby"
	LocationSearch_Type_Within     LocationDetect_Type = "within"
	LocationSearch_Type_Intersects LocationDetect_Type = "intersects"
)

// LocationDetect_Type
type LocationDetect_Type string

const (
	LocationDetect_Type_All     LocationDetect_Type = "inside,outside,enter,exit,cross"
	LocationDetect_Type_Inside  LocationDetect_Type = "inside"
	LocationDetect_Type_Outside LocationDetect_Type = "outside"
	LocationDetect_Type_Enter   LocationDetect_Type = "enter"
	LocationDetect_Type_Exit    LocationDetect_Type = "exit"
	LocationDetect_Type_Cross   LocationDetect_Type = "cross"
)

// LocationCommand_Type
type LocationCommand_Type string

const (
	LocationCommand_Type_All  LocationCommand_Type = "set,del,drop"
	LocationCommand_Type_Set  LocationCommand_Type = "set"
	LocationCommand_Type_Del  LocationCommand_Type = "del"
	LocationCommand_Type_Drop LocationCommand_Type = "drop"
)

// LocationObjectType
type LocationObject_Type string

const (
	LocationObject_Type_Point LocationObject_Type = "point"
)

// SearchAvailability
type NearbySearch_Availability string

const (
	NearbySearch_Availability_2   NearbySearch_Availability = "2"   // Busy
	NearbySearch_Availability_1   NearbySearch_Availability = "1"   // Available
	NearbySearch_Availability_0   NearbySearch_Availability = "0"   // Not Available
	NearbySearch_Availability_All NearbySearch_Availability = "all" // All
)

// SearchPriority
type NearbySearch_Priority string

const (
	NearbySearch_Priority_1   NearbySearch_Priority = "1"   // Priority
	NearbySearch_Priority_0   NearbySearch_Priority = "0"   // Not Priority
	NearbySearch_Priority_All NearbySearch_Priority = "all" // All
)

type DriverPriority int

const (
	DriverPriority_YES DriverPriority = 1
	DriverPriority_NO  DriverPriority = 0 // default status
)

// DriverStatus should match the proto message of driverstatuspoll
type DriverStatus int

const (
	DriverStatus_AVAILABLE    DriverStatus = 1
	DriverStatus_NOTAVAILABLE DriverStatus = 0 // default status
	DriverStatus_BUSY         DriverStatus = 2
)

func (status DriverStatus) String() string {
	// declare a map of int, string

	names := make(map[int]string)
	names[int(DriverStatus_AVAILABLE)] = "Available"
	names[int(DriverStatus_NOTAVAILABLE)] = "Not Available"
	names[int(DriverStatus_BUSY)] = "Busy"

	// prevent panicking in case of
	// `day` is out of range of Weekday
	if status < DriverStatus_NOTAVAILABLE || status > DriverStatus_BUSY {
		return "Unknown"
	}

	// return the name of a Drver Status
	// constant from the names array
	// above.
	return names[int(status)]
}

func (status DriverStatus) CanAcceptJob() bool {
	switch status {
	// status is Available:
	case DriverStatus_AVAILABLE:
		return true
	// else
	default:
		return false
	}
}
