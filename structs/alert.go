package structs

import (
	"time"

	"github.com/byuoitav/common/v2/events"
)

type RoomIssue struct {
	RoomIssueID string `json:"id"`

	events.BasicRoomInfo

	Severity AlertSeverity `json:"severity"`

	RoomTags []string `json:"room-tags"`

	AlertTypes      []AlertType     `json:"alert-types"`
	AlertCategories []AlertCategory `json:"alert-types"`

	SystemType string `json:"system-type"`

	Source string `json:"-"`

	Alerts map[string]Alert `json:"alerts"`

	//Editable fields
	IssueTags []string `json:"issue-tags"`

	IncidentID string `json:"incident-id"`

	Notes string `json:"notes"`

	Responders    []string  `json:"responders"`
	HelpSentAt    time.Time `json:"help-sent-at"`
	HelpArrivedAt time.Time `json:"help-arrived-at"`

	//resolution fields
	Resolved       bool           `json:"resolved"`
	ResolutionInfo ResolutionInfo `json:"resolution-info"`

	//notes-log isn't editable
	NotesLog []string `json:"notes-log"`
}

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
	SystemType string      `json:"system-type"`

	AlertStartTime      time.Time `json:"start-time"`
	AlertEndTime        time.Time `json:"end-time"`
	AlertLastUpdateTime time.Time `json:"update-time"`

	Active bool `json:"active"`

	AlertTags  []string `json:"alert-tags"`
	DeviceTags []string `json:"device-tags"`
	RoomTags   []string `json:"room-tags"`

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

func (r *RoomIssue) CalcualteTypeCategories() {
	r.AlertTypes = []AlertType{}

	r.AlertCategories = []AlertCategory{}

	for i := range r.Alerts {
		found := false
		for j := range r.AlertTypes {
			if r.AlertTypes[j] == r.Alerts[i].Type {
				found = true
				break
			}
		}
		if !found {
			r.AlertTypes = append(r.AlertTypes, r.Alerts[i].Type)
		}

		found = false
		for j := range r.AlertCategories {
			if r.AlertCategories[j] == r.Alerts[i].Category {
				found = true
				break
			}

		}
		if !found {
			r.AlertCategories = append(r.AlertCategories, r.Alerts[i].Category)
		}
	}
}
