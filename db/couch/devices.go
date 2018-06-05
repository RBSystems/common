package couch

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/common/structs"
)

func (c *CouchDB) GetDevice(id string) (structs.Device, error) {
	device, err := c.getDevice(id)
	return *device.Device, err
}

func (c *CouchDB) getDevice(id string) (device, error) {
	var toReturn device

	// get the device
	err := c.MakeRequest("GET", fmt.Sprintf("%s/%v", DEVICES, id), "", nil, &toReturn)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to get device %s: %s", id, err))
	}

	// get its device type
	toReturn.Type, err = c.GetDeviceType(toReturn.Type.ID)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to get device type (%s) to get device %s: %s", toReturn.Type.ID, id, err))
	}

	return toReturn, err
}

func (c *CouchDB) getDevicesByQuery(query IDPrefixQuery, includeType bool) ([]device, error) {
	var toReturn []device

	// marshal query
	b, err := json.Marshal(query)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to marshal devices query: %s", err))
	}

	// make query for devices
	var resp deviceQueryResponse
	err = c.MakeRequest("POST", fmt.Sprintf("%s/_find", DEVICES), "application/json", b, &resp)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to query devices: %s", err))
	}

	if includeType {
		// get all types
		types, err := c.GetAllDeviceTypes()
		if err != nil {
			return toReturn, errors.New(fmt.Sprintf("failed to get devices types for devices query:%s", err))
		}

		// make a map of type.ID -> type
		typesMap := make(map[string]structs.DeviceType)
		for _, t := range types {
			typesMap[t.ID] = t
		}

		// fill in device types
		for _, d := range resp.Docs {
			d.Type = typesMap[d.Type.ID]
		}
	}

	// return each document
	for _, doc := range resp.Docs {
		toReturn = append(toReturn, doc)
	}

	return toReturn, nil
}

func (c *CouchDB) GetAllDevices() ([]structs.Device, error) {
	var toReturn []structs.Device

	// create all device query
	var query IDPrefixQuery
	query.Selector.ID.GT = "\x00"
	query.Limit = 5000

	// query devices
	devices, err := c.getDevicesByQuery(query, false)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed getting all devices: %s", err))
	}

	// get the struct out of each device
	for _, device := range devices {
		toReturn = append(toReturn, *device.Device)
	}

	return toReturn, nil
}

func (c *CouchDB) GetDevicesByRoom(roomID string) ([]structs.Device, error) {
	var toReturn []structs.Device

	devices, err := c.getDevicesByRoom(roomID)
	if err != nil {
		return toReturn, err
	}

	for _, device := range devices {
		toReturn = append(toReturn, *device.Device)
	}

	return toReturn, nil
}

func (c *CouchDB) getDevicesByRoom(roomID string) ([]device, error) {
	var toReturn []device

	// create query
	var query IDPrefixQuery
	query.Selector.ID.GT = fmt.Sprintf("%v-", roomID)
	query.Selector.ID.LT = fmt.Sprintf("%v.", roomID)
	query.Limit = 1000

	// query devices
	toReturn, err := c.getDevicesByQuery(query, true)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed getting devices in room %s: %s", roomID, err))
	}

	return toReturn, nil
}

/*
Create Device. As amazing as it may seem, this fuction creates a device in the databse.

For a device to be created, it must contain the following attributes:

	1. A valid ID
		a. The room portion corresponds to an existing room
	2. A valid name
	3. A valid type
		a. Either the ID corresponds to an existing Type, or all elements are available to create a new type. Note that if the type ID matches, but the current type doesn't match the existing ID, the current type with that ID in the Database will NOT be overwritten.
	4. A valid Class
	5. One or more roles:
		a. A role must have a valid ID and Name

Ports must pass validation - criteria are covered in the CreateDeviceType function.
However in addition, if the port includes devices those devices must be valid devices

If a device is passed into the fuction with a valid 'rev' field, the current device with that ID will be overwritten.
`rev` must be omitted to create a new device.
*/
func (c *CouchDB) CreateDevice(toAdd structs.Device) (structs.Device, error) {
	var toReturn structs.Device

	// validate device struct
	err := toAdd.Validate()
	if err != nil {
		return toReturn, err
	}

	// validate room is real
	split := strings.Split(toAdd.ID, "-")
	roomID := split[0] + "-" + split[1]
	_, err = c.GetRoom(roomID)
	if err != nil {
		if _, ok := err.(*NotFound); ok {
			return toReturn, errors.New(fmt.Sprintf("unable to create device %s: room %s doesn't exist.", toAdd.ID, roomID))
		}

		return toReturn, errors.New(fmt.Sprintf("unable to validate device %s is in a real room: %s", toAdd.ID, err))
	}

	// validate device type
	deviceType, err := c.GetDeviceType(toAdd.Type.ID)
	if err != nil {
		if _, ok := err.(*NotFound); ok { // device type doesn't exist
			// try to create device type
			deviceType, err = c.CreateDeviceType(toAdd.Type)
			if err != nil {
				return toReturn, errors.New(fmt.Sprintf("attempting to create a device with a non-existant device type, but not enough information is included to create the type. (error: %s)", err))
			}
		} else { // unknown error getting device type
			return toReturn, errors.New(fmt.Sprintf("unable to validate if device type %s exists or not: %s", toAdd.Type.ID, err))
		}
	}

	// the device document should only include the type ID
	toAdd.Type = structs.DeviceType{ID: deviceType.ID}

	// check that each of the ports are valid
	for _, port := range toAdd.Ports {
		if err = c.checkPort(port); err != nil {
			return toReturn, errors.New(fmt.Sprintf("unable to create device: %s", err))
		}
	}

	// marshal the device
	b, err := json.Marshal(toAdd)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to marshal device: %s", err))
	}

	// post up device
	var resp CouchUpsertResponse
	err = c.MakeRequest("POST", fmt.Sprintf("%v", DEVICES), "application/json", b, &resp)
	if err != nil {
		if _, ok := err.(*Conflict); ok { // device with same id already in database
			return toReturn, errors.New(fmt.Sprintf("unable to create device, because it already exists. error: %s", err))
		}

		return toReturn, errors.New(fmt.Sprintf("unknown error creating device %s: %s", toAdd.ID, err))
	}

	// return the device that is in the database
	toReturn, err = c.GetDevice(toAdd.ID)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("unable to get device %s: %s", toAdd.ID, err))
	}

	return toReturn, nil
}

func (c *CouchDB) DeleteDevice(id string) error {
	// get the device to delete
	device, err := c.getDevice(id)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to get device %s to delete: %s", id, err))
	}

	// delete the device
	err = c.MakeRequest("DELETE", fmt.Sprintf("%v/%v?rev=%v", DEVICES, device.ID, device.Rev), "", nil, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to delete device %s: %s", id, err))
	}

	return nil
}

// TODO make this actually update, as opposed to deleting/creating.
//	 this way you don't have to post up a full document
// 	 probably need to do this for all of the update functions
func (c *CouchDB) UpdateDevice(id string, device structs.Device) (structs.Device, error) {
	var toReturn structs.Device

	// validate the new struct
	err := device.Validate()
	if err != nil {
		return toReturn, err
	}

	// delete the old struct
	err = c.DeleteDevice(id)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to update device %s: %s", id, err))
	}

	// create new version of device
	toReturn, err = c.CreateDevice(device)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to update device %s: %s", device.ID, err))
	}

	return toReturn, err
}

func (c *CouchDB) checkPort(p structs.Port) error {
	// check source port
	if len(p.SourceDevice) > 0 {
		if _, err := c.GetDevice(p.SourceDevice); err != nil {
			return errors.New(fmt.Sprintf("invalid port %v. source device %v doesn't exist. Create it before adding it to a port", p.ID, p.SourceDevice))
		}
	}

	// check desitnation port
	if len(p.DestinationDevice) > 0 {
		if _, err := c.GetDevice(p.DestinationDevice); err != nil {
			return errors.New(fmt.Sprintf("invalid port %v. destination device %v doesn't exist. Create it before adding it to a port", p.ID, p.DestinationDevice))
		}
	}

	return nil
}

func (c *CouchDB) GetDevicesByRoomAndRole(roomID, role string) ([]structs.Device, error) {
	toReturn := []structs.Device{}

	// get all devices in room
	devs, err := c.GetDevicesByRoom(roomID)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to get devices by room and role: %s", err))
	}

	// go through the devices and check if they have the role indicated
	for _, d := range devs {
		if structs.HasRole(d, role) {
			toReturn = append(toReturn, d)
		}
	}

	return toReturn, nil
}

// TODO could actually use a query to be faster
func (c *CouchDB) GetDevicesByType(deviceType string) ([]structs.Device, error) {
	var toReturn []structs.Device

	// get all devices
	devs, err := c.GetAllDevices()
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to get devices by type: %s", err))
	}

	// filter for ones that have correct type
	for _, d := range devs {
		if strings.EqualFold(d.Type.ID, deviceType) {
			toReturn = append(toReturn, d)
		}
	}

	return toReturn, nil
}

// TODO a real query would probably be faster again
func (c *CouchDB) GetDevicesByRoleAndType(role, deviceType string) ([]structs.Device, error) {
	var toReturn []structs.Device

	// get all devices
	devs, err := c.GetAllDevices()
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to get devices by role and type: %s", err))
	}

	// filter for ones that have the role and type
	for _, d := range devs {
		if structs.HasRole(d, role) && strings.EqualFold(d.Type.ID, deviceType) {
			toReturn = append(toReturn, d)
		}
	}

	return toReturn, nil
}
