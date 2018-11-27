package structs

// HardwareInfo contains the common information for device hardware information
type HardwareInfo struct {
	Hostname        string      `json:"hostname,omitempty"`
	ModelName       string      `json:"model_name,omitempty"`
	SerialNumber    string      `json:"serial_number,omitempty"`
	FirmwareVersion interface{} `json:"firmware_version,omitempty"`
	NetworkInfo     NetworkInfo `json:"network_information,omitempty"`
	FilterStatus    string      `json:"filter_status,omitempty"`
	WarningStatus   []string    `json:"warning_status,omitempty"`
	ErrorStatus     []string    `json:"error_status,omitempty"`
	PowerStatus     string      `json:"power_status,omitempty"`
	TimerInfo       interface{} `json:"timer_info,omitempty"`
}

// NetworkInfo contains the network information for the device
type NetworkInfo struct {
	IPAddress  string   `json:"ip_address,omitempty"`
	MACAddress string   `json:"mac_address,omitempty"`
	DNS        []string `json:"dns,omitempty"`
}
