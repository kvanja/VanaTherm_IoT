package dbstructs

import (
	"fmt"
	"strconv"
)

type Buildings struct {
	ID           uint     `json:"ID"`
	Name         string   `json:"NAME"`
	Street       string   `json:"STREET"`
	StreetNumber uint     `json:"STREET_NUMBER"`
	Country      string   `json:"COUNTRY"`
	City         string   `json:"CITY"`
	ZipCode      uint     `json:"ZIP_CODE"`
	Location     string   `json:"LOCATION"`
	Created      string   `json:"CREATED"`
	Updated      string   `json:"UPDATED"`
	CreatedBy    uint     `json:"CREATED_BY"`
	UpdatedBy    uint     `json:"UPDATED_BY"`
	Floors       []Floors `json:"FLOORS"`
}

type Floors struct {
	ID         uint    `json:"ID"`
	BuildingId uint    `json:"BUILDING_ID"`
	Floor      int     `json:"FLOOR"`
	Devices    uint    `json:"DEVICES"`
	Created    string  `json:"CREATED"`
	Updated    string  `json:"UPDATED"`
	CreatedBy  uint    `json:"CREATED_BY"`
	UpdatedBy  uint    `json:"UPDATED_BY"`
	Rooms      []Rooms `json:"ROOMS"`
}

func (floor *Floors) FloorExistsError() string {
	var floorNum int
	var floorMod100 int = floor.Floor % 100
	if floorMod100 == 11 || floorMod100 == 12 || floorMod100 == 13 {
		floorNum = floorMod100
	} else {
		floorNum = floor.Floor % 10
	}

	var ordinalIndicator string
	switch floorNum {
	case 1:
		ordinalIndicator = "st "
	case 2:
		ordinalIndicator = "nd "
	case 3:
		ordinalIndicator = "rd "
	default:
		ordinalIndicator = "th "
	}

	var floorOrdinal string = strconv.Itoa(floor.Floor) + ordinalIndicator

	return fmt.Sprint("Building with ", floorOrdinal, "floor already exists ")
}

type Rooms struct {
	ID         uint      `json:"ID"`
	BuildingId uint      `json:"BUILDING_ID"`
	FloorId    uint      `json:"FLOOR_ID"`
	RoomName   string    `json:"ROOM_NAME"`
	Devices    []Devices `json:"DEVICES"`
	Created    string    `json:"CREATED"`
	Updated    string    `json:"UPDATED"`
	CreatedBy  uint      `json:"CREATED_BY"`
	UpdatedBy  uint      `json:"UPDATED_BY"`
}

type Devices struct {
	ID              uint   `json:"ID"`
	RoomId          uint   `json:"ROOM_ID"`
	IsControlDevice bool   `json:"IS_CONTROL_DEVICE"`
	IsControlOn     bool   `json:"IS_CONTROL_ON"`
	IsDeviceActive  bool   `json:"IS_DEVICE_ACTIVE"`
	Created         string `json:"CREATED"`
	Updated         string `json:"UPDATED"`
	CreatedBy       uint   `json:"CREATED_BY"`
	UpdatedBy       uint   `json:"UPDATED_BY"`
}

type UserDevices struct {
	UserID   uint `json:"USER_ID"`
	DeviceID uint `json:"DEVICE_ID"`
}

type UserBuildingsAndDevices struct {
	Buildings []Buildings `json:"BUILDINGS"`
}

func (ubdv *UserBuildingsAndDevices) FindBuilding(id uint) []int {
	for i, el := range ubdv.Buildings {
		if el.ID == id {
			return []int{i, int(id)}
		}
	}
	return []int{0, 0}
}

func (ubdv *UserBuildingsAndDevices) FindFloor(id uint) []int {
	for _, building := range ubdv.Buildings {
		for i, floor := range building.Floors {
			if floor.ID == id {
				return []int{i, int(id)}
			}
		}
	}
	return []int{0, 0}
}

func (ubdv *UserBuildingsAndDevices) FindRoom(id uint) []int {
	for _, building := range ubdv.Buildings {
		for _, floor := range building.Floors {
			for i, room := range floor.Rooms {
				if room.ID == id {
					return []int{i, int(id)}
				}
			}
		}
	}
	return []int{0, 0}
}

func (ubdv *UserBuildingsAndDevices) FindDevice(id uint) []int {
	for _, building := range ubdv.Buildings {
		for _, floor := range building.Floors {
			for _, room := range floor.Rooms {
				for i, device := range room.Devices {
					if device.ID == id {
						return []int{i, int(id)}
					}
				}
			}
		}
	}
	return []int{0, 0}
}

type DeviceReadData struct {
	Temperature float32 `json:"temperature"`
	TempUnit    string  `json:"tempUnit"`
	Humidity    float32 `json:"humidity"`
}

type HeatingHistory struct {
	ID                    uint    `json:"id"`
	UserId                uint    `json:"userId"`
	DeviceID              uint    `json:"deviceId"`
	TempMeasurementUnitId uint    `json:"tempMeasurementUnitId"`
	Date                  string  `json:"date"`
	TimeFrom              string  `json:"timeFrom"`
	TimeTo                string  `json:"timeTo"`
	CurrentTemperature    float32 `json:"currentTemperature"`
	DesiredTemperature    int     `json:"desiredTemperature"`
	Created               string  `json:"created"`
	Updated               string  `json:"updated"`
	CreatedBy             uint    `json:"createdBy"`
	UpdatedBy             uint    `json:"updatedBy"`
}

type UserDevicesHeatingHistory struct {
	BuildingName       string  `json:"buildingName"`
	City               string  `json:"city"`
	Street             string  `json:"street"`
	Floor              int     `json:"floor"`
	RoomName           string  `json:"roomName"`
	DeviceID           int     `json:"deviceId"`
	Date               string  `json:"date"`
	TimeFrom           string  `json:"timeFrom"`
	TimeTo             string  `json:"timeTo"`
	CurrentTemperature float32 `json:"currentTemperature"`
	DesiredTemp        float32 `json:"desiredTemperature"`
	Unit               string  `json:"unit"`
}

type UserDevicesHeatingHistories struct {
	UserDevicesHeatingHistory []UserDevicesHeatingHistory `json:"userDevicesHeatingHistory"`
}
