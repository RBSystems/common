package structs

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Device - a representation of a device involved in a TEC Pi system.
type Device struct {
	ID          string     `json:"_id"`
	Name        string     `json:"name"`
	Address     string     `json:"address"`
	Description string     `json:"description"`
	DisplayName string     `json:"display_name"`
	Type        DeviceType `json:"type,omitempty"`
	Roles       []Role     `json:"roles"`
	Ports       []Port     `json:"ports"`
	Tags        []string   `json:"tags,omitempty"`
}

// DeviceIDValidationRegex is our regular expression for validating the correct naming scheme.
var deviceIDValidationRegex = regexp.MustCompile(`([A-z,0-9]{2,}-[A-z,0-9]+)-[A-z]+[0-9]+`)

// IsDeviceIDValid takes a device id and tells you whether or not it is valid.
func IsDeviceIDValid(id string) bool {
	reg := deviceIDValidationRegex.Copy()

	vals := reg.FindStringSubmatch(id)
	if len(vals) == 0 {
		return false
	}
	return true
}

// Validate checks to see if the device's information is valid or not.
func (d *Device) Validate() error {
	vals := deviceIDValidationRegex.FindStringSubmatch(d.ID)
	if len(vals) == 0 {
		return errors.New("invalid device: inproper id. must match `([A-z,0-9]{2,}-[A-z,0-9]+)-[A-z]+[0-9]+`")
	}

	if len(d.Name) < 2 {
		return errors.New("invalid device: name must be at least 3 characters long")
	}

	// validate device type
	if err := d.Type.Validate(false); err != nil {
		return fmt.Errorf("invalid device: %s", err)
	}

	// validate roles
	if len(d.Roles) == 0 {
		return errors.New("invalid device: must include at least 1 role")
	}
	for _, role := range d.Roles {
		if err := role.Validate(); err != nil {
			return fmt.Errorf("invalid device: %s", err)
		}
	}

	// validate ports
	for _, port := range d.Ports {
		if err := port.Validate(); err != nil {
			return fmt.Errorf("invalid device: %s", err)
		}
	}

	return nil
}

// GetDeviceRoomID returns the room ID portion of the device ID.
func (d *Device) GetDeviceRoomID() string {
	idParts := strings.Split(d.ID, "-")
	roomID := fmt.Sprintf("%s-%s", idParts[0], idParts[1])
	return roomID
}

// GetCommandByName searches for a specific command and returns it if found.
func (d *Device) GetCommandByName(port string) Command {
	for _, c := range d.Type.Commands {
		if c.ID == port {
			return c
		}
	}

	// No command found.
	return Command{}
}

// DeviceType - a representation of a type (or category) of devices.
type DeviceType struct {
	ID          string       `json:"_id"`
	Description string       `json:"description,omitempty"`
	DisplayName string       `json:"display_name,omitempty"`
	Input       bool         `json:"input,omitempty"`
	Output      bool         `json:"output,omitempty"`
	Source      bool         `json:"source,omitempty"`
	Destination bool         `json:"destination,omitempty"`
	Roles       []Role       `json:"roles,omitempty"`
	Ports       []Port       `json:"ports,omitempty"`
	PowerStates []PowerState `json:"power_states,omitempty"`
	Commands    []Command    `json:"commands,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
}

// Validate checks to make sure that the values of the DeviceType are valid.
func (dt *DeviceType) Validate(deepCheck bool) error {
	if len(dt.ID) == 0 {
		return errors.New("invalid device type: missing id")
	}

	if deepCheck {
		// check all of the ports
		for _, port := range dt.Ports {
			if err := port.Validate(); err != nil {
				return fmt.Errorf("invalid device type: %s", err)
			}
		}

		// check all of the commands
		for _, command := range dt.Commands {
			if err := command.Validate(); err != nil {
				return fmt.Errorf("invalid device type: %s", err)
			}
		}
	}
	return nil
}

// PowerState - a representation of a device's power state.
type PowerState struct {
	ID          string   `json:"_id"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
}

// Validate checks to make sure that the PowerState's values are valid.
func (ps *PowerState) Validate() error {
	if len(ps.ID) < 3 {
		return errors.New("invalid power state: id must be at least 3 characters long")
	}
	return nil
}

// Port - a representation of an input/output port on a device.
type Port struct {
	ID                string   `json:"_id"`
	FriendlyName      string   `json:"friendly_name,omitempty"`
	PortType          string   `json:"port_type,omitempty"`
	SourceDevice      string   `json:"source_device,omitempty"`
	DestinationDevice string   `json:"destination_device,omitempty"`
	Description       string   `json:"description,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

// Validate checks to make sure that the Port's values are valid.
func (p *Port) Validate() error {
	if len(p.ID) < 1 {
		return errors.New("invalid port: id must be at least 3 characters long")
	}
	return nil
}

// Role - a representation of a role that a device plays in the overall system.
type Role struct {
	ID          string   `json:"_id"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// Validate checks to make sure that the Role's values are valid.
func (r *Role) Validate() error {
	if len(r.ID) < 3 {
		return errors.New("invalid role: id must at least 3 characters long")
	}
	return nil
}

// Command - a representation of an API command to be executed.
type Command struct {
	ID           string       `json:"_id"`
	Description  string       `json:"description"`
	Microservice Microservice `json:"microservice"`
	Endpoint     Endpoint     `json:"endpoint"`
	Priority     int          `json:"priority"`
	Tags         []string     `json:"tags,omitempty"`
}

// Validate checks to make sure that the Command's values are valid.
func (c *Command) Validate() error {
	if len(c.ID) < 3 {
		return errors.New("invalid command: id must be at least 3 characters long")
	}

	// validate microservice
	if err := c.Microservice.Validate(); err != nil {
		return fmt.Errorf("invalid command: %s", err)
	}

	// validate endpoint
	if err := c.Endpoint.Validate(); err != nil {
		return fmt.Errorf("invalid command: %s", err)
	}
	return nil
}

// Microservice - a representation of a microservice in our API.
type Microservice struct {
	ID          string   `json:"_id"`
	Description string   `json:"description"`
	Address     string   `json:"address"`
	Tags        []string `json:"tags,omitempty"`
}

// Validate checks to make sure that the Microservice's values are valid.
func (m *Microservice) Validate() error {
	if len(m.ID) < 3 {
		return errors.New("invalid microservice: id must be at least 3 characters long")
	}

	// validate address
	if _, err := url.ParseRequestURI(m.Address); err != nil {
		return fmt.Errorf("invalid microservice: %s", err)
	}
	return nil
}

// Endpoint - a representation of an API endpoint.
type Endpoint struct {
	ID          string   `json:"_id"`
	Description string   `json:"description"`
	Path        string   `json:"path"`
	Tags        []string `json:"tags,omitempty"`
}

// Validate checks to make sure that the Endpoint's values are valid.
func (e *Endpoint) Validate() error {
	if len(e.ID) < 3 {
		return errors.New("invalid endpoint: id must be at least 3 characters long")
	}

	// validate path
	if _, err := url.ParseRequestURI(e.Path); err != nil {
		return fmt.Errorf("invalid endpoint: %s", err)
	}
	return nil
}

// HasRole checks to see if the given device has the given role.
func HasRole(device Device, role string) bool {
	role = strings.ToLower(role)
	for i := range device.Roles {
		if strings.EqualFold(strings.ToLower(device.Roles[i].ID), role) {
			return true
		}
	}
	return false
}

// HasRole checks to see if the given device has the given role.
func (d *Device) HasRole(role string) bool {
	role = strings.ToLower(role)
	for i := range d.Roles {
		if strings.EqualFold(strings.ToLower(d.Roles[i].ID), role) {
			return true
		}
	}
	return false
}
