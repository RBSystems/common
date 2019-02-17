package structs

import (
	"time"

	"github.com/byuoitav/common/v2/events"
)

// Alert is a struct that contains the information regarding an alerting event.
type Alert struct {
	events.BasicDeviceInfo

	AlertID string `json:"id,omitempty"`

	Type     AlertType     `json:"type"`
	Category AlertCategory `json:"category"`
	Severity AlertSeverity `json:"severity"`

	Message    string      `json:"message"`
	MessageLog []string    `json:"message-log"`
	Data       interface{} `json:"data,omitempty"`
	IncidentID string      `json:"incident-id"`
	SystemType string      `json:"system-type"`

	Notes    []string `json:"notes"`
	LastNote string   `json:"last-notes"`

	AlertStartTime      time.Time `json:"start-time"`
	AlertEndTime        time.Time `json:"end-time"`
	AlertLastUpdateTime time.Time `json:"update-time"`
	Active              bool      `json:"active"`

	Resolved       bool           `json:"resolved"`
	Responders     []string       `json:"responders"`
	HelpSentAt     time.Time      `json:"help-sent-at"`
	HelpArrivedAt  time.Time      `json:"help-arrived-at"`
	ResolutionInfo ResolutionInfo `json:"resolution-info"`

	AlertTags  []string `json:"alert-tags"`
	RoomTags   []string `json:"room-tags"`
	DeviceTags []string `json:"device-tags"`

	Source string `json:"-"`
}

// AlertType is an enum of the different types of alerts
type AlertType string

const (
	Communication AlertType = "communication"
	Heartbeat     AlertType = "heartbeat"
)

// AlertCategory is an enum of the different categories of alerts
type AlertCategory string

// Here is a list of AlertCategory
const (
	System AlertCategory = "system"
	User   AlertCategory = "user"
)

// AlertSeverity is an enum of the different levels of severity for alerts
type AlertSeverity string

// Here is a list of AlertSeverities
const (
	Critical AlertSeverity = "critical"
	Warning  AlertSeverity = "warning"
	Low      AlertSeverity = "low"
)

// ResolutionInfo is a struct that contains the information about the resolution of the alert
type ResolutionInfo struct {
	Code           string    `json:"resolution-code"`
	Notes          string    `json:"notes"`
	ResolvedAt     time.Time `json:"resolved-at"`
	ResolutionHash string    `json:"resolution-hash"`
}

func ContainsAllTags(tagList []string, tags ...string) bool {
	for i := range tags {
		hasTag := false

		for j := range tagList {
			if tagList[j] == tags[i] {
				hasTag = true
				continue
			}
		}

		if !hasTag {
			return false
		}
	}

	return true
}

func AddToTags(tagList []string, tags ...string) []string {
	for _, t := range tags {
		if !ContainsAllTags(tagList, t) {
			tagList = append(tagList, t)
		}
	}
	return tagList
}

func ContainsAnyTags(tagList []string, tags ...string) bool {
	for i := range tags {
		for j := range tagList {
			if tagList[j] == tags[i] {
				return true
			}
		}
	}

	return false
}
