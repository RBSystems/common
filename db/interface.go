package db

import (
	"os"

	"github.com/byuoitav/common/db/couch"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/structs"
)

type DB interface {
	/* crud functions */
	// building
	CreateBuilding(building structs.Building) (structs.Building, error)
	GetBuilding(id string) (structs.Building, error)
	UpdateBuilding(id string, building structs.Building) (structs.Building, error)
	DeleteBuilding(id string) error

	// room
	CreateRoom(room structs.Room) (structs.Room, error)
	GetRoom(id string) (structs.Room, error)
	UpdateRoom(id string, room structs.Room) (structs.Room, error)
	DeleteRoom(id string) error

	// device
	CreateDevice(device structs.Device) (structs.Device, error)
	GetDevice(id string) (structs.Device, error)
	UpdateDevice(id string, device structs.Device) (structs.Device, error)
	DeleteDevice(id string) error

	// device type
	CreateDeviceType(dt structs.DeviceType) (structs.DeviceType, error)
	GetDeviceType(id string) (structs.DeviceType, error)
	UpdateDeviceType(id string, dt structs.DeviceType) (structs.DeviceType, error)
	DeleteDeviceType(id string) error

	// room configuration
	CreateRoomConfiguration(rc structs.RoomConfiguration) (structs.RoomConfiguration, error)
	GetRoomConfiguration(id string) (structs.RoomConfiguration, error)
	UpdateRoomConfiguration(id string, rc structs.RoomConfiguration) (structs.RoomConfiguration, error)
	DeleteRoomConfiguration(id string) error

	// ui configs
	CreateUIConfig(roomID string, ui structs.UIConfig) (structs.UIConfig, error)
	GetUIConfig(roomID string) (structs.UIConfig, error)
	UpdateUIConfig(id string, ui structs.UIConfig) (structs.UIConfig, error)
	DeleteUIConfig(id string) error

	/* bulk functions */
	GetAllBuildings() ([]structs.Building, error)
	GetAllRooms() ([]structs.Room, error)
	GetAllDevices() ([]structs.Device, error)
	GetAllDeviceTypes() ([]structs.DeviceType, error)
	GetAllRoomConfigurations() ([]structs.RoomConfiguration, error)
	CreateBulkDevices([]structs.Device) []structs.BulkUpdateResponse // TODO change the response struct

	/* Specialty functions */
	GetDevicesByRoom(roomID string) ([]structs.Device, error)
	GetDevicesByRoomAndType(roomID, typeID string) ([]structs.Device, error)
	GetDevicesByRoomAndRole(roomID, roleID string) ([]structs.Device, error)
	GetDevicesByRoleAndType(roleID, typeID string) ([]structs.Device, *nerr.E)
	GetDevicesByRoleAndTypeAndDesignation(roleID, typeID, designation string) ([]structs.Device, *nerr.E)

	GetRoomsByBuilding(id string) ([]structs.Room, error)
	GetRoomsByDesignation(designation string) ([]structs.Room, *nerr.E)

	/* Options Functions */
	GetTemplate(id string) (structs.UIConfig, error)
	GetAllTemplates() ([]structs.Template, error)
	UpdateTemplate(id string, newTemp structs.UIConfig) (structs.UIConfig, error)
	GetIcons() ([]string, error)
	UpdateIcons(iconList []string) ([]string, error)
	GetDeviceRoles() ([]structs.Role, error)
	UpdateDeviceRoles(roles []structs.Role) ([]structs.Role, error)
	GetRoomDesignations() ([]string, error)
	UpdateRoomDesignations(desigs []string) ([]string, error)
	GetTags() ([]string, error)
	UpdateTags(newTags []string) ([]string, error)

	GetAuth() (structs.Auth, error)

	//Get the state (replication/readiness) of the database
	GetStatus() (string, error)
}

var address string
var username string
var password string

var database DB

func init() {
	address = os.Getenv("DB_ADDRESS")
	username = os.Getenv("DB_USERNAME")
	password = os.Getenv("DB_PASSWORD")

	if len(address) == 0 {
		log.L.Errorf("DB_ADDRESS is not set. Failing...")
	}
}

// GetDB returns the instance of the database to use.
func GetDB() DB {
	// TODO add logic to "pick" which db to create
	if database == nil {
		database = couch.NewDB(address, username, password)
	}

	return database
}

// GetDBWithCustomAuth returns an instance of the database with a custom authentication
func GetDBWithCustomAuth(address, username, password string) DB {
	return couch.NewDB(address, username, password)
}
