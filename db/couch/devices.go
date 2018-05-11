package couch

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/common/structs"
)

// var DeviceValidationRegex = regexp.MustCompile(`([A-z,0-9]{2,}-[A-z,0-9]+)-[A-z]+[0-9]+`)

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
			return toReturn, errors.New(fmt.Sprintf("failed to get devices types for devices query: ", err))
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

	// the device document shoudl only include the type ID
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

	/*
		//return the created room
		toAdd, err = c.GetDevice(toAdd.ID)
		if err != nil {
			c.lde(fmt.Sprintf("There was a problem getting the newly created room: %v", err.Error()))
		}

		c.log.Debug("Done creating device.")
		return toAdd, nil
	*/
	return toReturn, nil
}

/*
//log device error
//alias to help cut down on cruft
func (c *CouchDB) lde(msg string) (dev structs.Device, err error) {
	c.log.Warn(msg)
	err = errors.New(msg)
	return
}
*/

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

	devs, err := c.GetDevicesByRoom(roomID)
	if err != nil {
		msg := fmt.Sprintf("Couldn't get devices for filtering: %v", err.Error())
		c.log.Warn(msg)
		return toReturn, errors.New(msg)
	}

	//go through the devices and check if they have the role indicated
	for _, d := range devs {
		if structs.HasRole(d, role) {
			toReturn = append(toReturn, d)
		}
	}

	return toReturn, nil
}

func (c *CouchDB) GetDevicesByRoleAndType(role, dtype string) ([]structs.Device, error) {
	var toReturn []structs.Device

	devs, err := c.GetAllDevices()
	if err != nil {
		msg := fmt.Sprintf("unable to get device list: %v", err.Error())
		c.log.Warn(msg)
		return toReturn, err
	}

	for _, d := range devs {
		if structs.HasRole(d, role) && strings.EqualFold(d.Type.ID, dtype) {
			toReturn = append(toReturn, d)
		}
	}

	return toReturn, nil
}

func (c *CouchDB) DeleteDevice(id string) error {
	c.log.Debugf("[%s] Deleting device", id)

	device, err := c.getDevice(id)
	if err != nil {
		msg := fmt.Sprintf("[%s] error looking for device to delete: %s", id, err.Error())
		c.log.Warn(msg)
		return errors.New(msg)
	}

	err = c.MakeRequest("DELETE", fmt.Sprintf("devices/%s?rev=%v", device.ID, device.Rev), "", nil, nil)
	if err != nil {
		msg := fmt.Sprintf("[%s] error deleting device: %s", id, err.Error())
		c.log.Warn(msg)
		return errors.New(msg)
	}

	return nil
}

func (c *CouchDB) UpdateDevice(id string, device structs.Device) (structs.Device, error) {
	var toReturn structs.Device

	b, err := json.Marshal(device)
	if err != nil {
		msg := fmt.Sprintf("there was a problem marshalling the query: %s", err)
		c.log.Warnf(msg)
		return toReturn, errors.New(msg)
	}

	dev, err := c.getDevice(id)
	if err != nil {
		msg := fmt.Sprintf("error getting the device to delete: %s", err)
		c.log.Warnf(msg)
		return toReturn, errors.New(msg)
	}

	err = c.MakeRequest("PUT", fmt.Sprintf("devices/%s?rev=%v", device.ID, dev.Rev), "application/json", b, &toReturn)
	if err != nil {
		msg := fmt.Sprintf("error updating the device %s: %s", device.ID, err)
		c.log.Warn(msg)
		return toReturn, errors.New(msg)
	}

	if id != device.ID {
		// delete the old document
		err = c.DeleteDevice(id)
		if err != nil {
			msg := fmt.Sprintf("error deleting the old device %s: %s", id, err)
			c.log.Warn(msg)
			return toReturn, errors.New(msg)
		}
	}

	return toReturn, nil
}
