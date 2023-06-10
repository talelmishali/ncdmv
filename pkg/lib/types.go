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
	AppointmentTypeDriverLicenseDuplicate AppointmentType = 2
	AppointmentTypeDriverLicenseRenewal   AppointmentType = 3
	AppointmentTypeIdCard                 AppointmentType = 5
	AppointmentTypeKnowledgeTest          AppointmentType = 6
	AppointmentTypeMotorcycleSkillsTest   AppointmentType = 8
	AppointmentTypePermit                 AppointmentType = 9
)

func (a AppointmentType) String() string {
	switch a {
	case AppointmentTypeInvalid:
		return "invalid"
	case AppointmentTypeDriverLicense:
		return "driver-license"
	case AppointmentTypeDriverLicenseDuplicate:
		return "driver-license-duplicate"
	case AppointmentTypeDriverLicenseRenewal:
		return "driver-license-renewal"
	case AppointmentTypeIdCard:
		return "id-card"
	case AppointmentTypeKnowledgeTest:
		return "knowledge-test"
	case AppointmentTypeMotorcycleSkillsTest:
		return "motorcycle-skills-test"
	case AppointmentTypePermit:
		return "permit"
	}
	panic("unreachable: invalid AppointmentType")
}

var appointmentTypeMap map[string]AppointmentType = map[string]AppointmentType{
	AppointmentTypeDriverLicense.String():          AppointmentTypeDriverLicense,
	AppointmentTypeDriverLicenseDuplicate.String(): AppointmentTypeDriverLicenseDuplicate,
	AppointmentTypeDriverLicenseRenewal.String():   AppointmentTypeDriverLicenseRenewal,
	AppointmentTypeIdCard.String():                 AppointmentTypeIdCard,
	AppointmentTypeKnowledgeTest.String():          AppointmentTypeKnowledgeTest,
	AppointmentTypeMotorcycleSkillsTest.String():   AppointmentTypeMotorcycleSkillsTest,
	AppointmentTypePermit.String():                 AppointmentTypePermit,
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
	// Values taken from the "data-id" attribute set on each location div.
	LocationInvalid      Location = iota
	LocationAhoskie      Location = 99
	LocationCary         Location = 66
	LocationDurhamEast   Location = 47
	LocationDurhamSouth  Location = 80
	LocationFuquayVarina Location = 38
	LocationGarner       Location = 69
	LocationHillsborough Location = 52
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
	case LocationFuquayVarina:
		return "fuquay-varina"
	case LocationGarner:
		return "garner"
	case LocationHillsborough:
		return "hillsborough"
	case LocationRaleighEast:
		return "raleigh-east"
	case LocationRaleighNorth:
		return "raleigh-north"
	case LocationRaleighWest:
		return "raleigh-west"
	}
	panic("unreachable: invalid Location")
}

var locationMap map[string]Location = map[string]Location{
	LocationAhoskie.String():      LocationAhoskie,
	LocationCary.String():         LocationCary,
	LocationDurhamEast.String():   LocationDurhamEast,
	LocationDurhamSouth.String():  LocationDurhamSouth,
	LocationFuquayVarina.String(): LocationFuquayVarina,
	LocationGarner.String():       LocationGarner,
	LocationHillsborough.String(): LocationHillsborough,
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
