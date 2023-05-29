package lib

import (
	"fmt"
)

const (
	apptTypeSelector = `div.QflowObjectItem[data-id="%d"]`
	locationSelector = apptTypeSelector
)

func mapToKeys[K comparable, V any](m map[K]V) []string {
	var keys []string
	for k := range m {
		keys = append(keys, fmt.Sprintf("%v", k))
	}
	return keys
}

// AppointmentType represents the type of appointment.
// The value is the index of the box in the UI (see: "data-id").
type AppointmentType int

func (a AppointmentType) ToSelector() string {
	return fmt.Sprintf(apptTypeSelector, a)
}

const (
	AppointmentTypeInvalid                AppointmentType = iota
	AppointmentTypeDriverLicense          AppointmentType = 1
	AppointmentTypeDriverLicenseDuplicate                 = 2
	AppointmentTypeDriverLicenseRenewal                   = 3
	AppointmentTypePermit                                 = 9
)

var appointmentTypeMap map[string]AppointmentType = map[string]AppointmentType{
	"license":           AppointmentTypeDriverLicense,
	"license-duplicate": AppointmentTypeDriverLicenseDuplicate,
	"license-renewal":   AppointmentTypeDriverLicenseRenewal,
	"permit":            AppointmentTypePermit,
}

func StringToAppointmentType(k string) AppointmentType {
	if v, ok := appointmentTypeMap[k]; !ok {
		return AppointmentTypeInvalid
	} else {
		return v
	}
}

func ValidApptTypes() []string {
	return mapToKeys(appointmentTypeMap)
}

type Location int

const (
	LocationInvalid      Location = iota
	LocationAhoskie      Location = 99
	LocationCary         Location = 66
	LocationDurhamEast   Location = 47
	LocationDurhamSouth  Location = 80
	LocationRaleighEast  Location = 181
	LocationRaleighNorth Location = 10
	LocationRaleighWest  Location = 9
)

func (l Location) ToSelector() string {
	return fmt.Sprintf(locationSelector, l)
}

func (l Location) String() string {
	switch l {
	case LocationAhoskie:
		return "ahoskie"
	case LocationCary:
		return "cary"
	case LocationDurhamEast:
		return "durham-east"
	case LocationDurhamSouth:
		return "durham-south"
	case LocationRaleighEast:
		return "raleigh-east"
	case LocationRaleighNorth:
		return "raleigh-north"
	case LocationRaleighWest:
		return "raleigh-west"
	}
	return ""
}

var locationMap map[string]Location = map[string]Location{
	LocationAhoskie.String():      LocationAhoskie,
	LocationCary.String():         LocationCary,
	LocationDurhamEast.String():   LocationDurhamEast,
	LocationDurhamSouth.String():  LocationDurhamSouth,
	LocationRaleighEast.String():  LocationRaleighEast,
	LocationRaleighNorth.String(): LocationRaleighNorth,
	LocationRaleighWest.String():  LocationRaleighWest,
}

func StringToLocation(k string) Location {
	if v, ok := locationMap[k]; !ok {
		return LocationInvalid
	} else {
		return v
	}
}

func ValidLocations() []string {
	return mapToKeys(locationMap)
}
