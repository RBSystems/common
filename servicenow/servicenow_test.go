package servicenow

import (
	"testing"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/common/v2/events"
)

func TestServiceNow(t *testing.T) {
	log.SetLevel("debug")

	TestAlert := structs.Alert{
		BasicDeviceInfo: events.BasicDeviceInfo{
			BasicRoomInfo: events.BasicRoomInfo{
				BuildingID: "ITB",
				RoomID:     "ITB-1108",
			},
			DeviceID: "ITB-1108-CP1",
		},
		AlertID:    "ITB-1108-CP1^System Communication Error^System^Critical",
		Type:       "System Communication Error",
		Category:   "System",
		Severity:   "Critical",
		Message:    "ITB-1108-CP1 has not reported any state since 2019-02-18 13:38:18.639429878 -0700 MST",
		SystemType: "pi",
		Data:       "Data of the event goes here",
	}

	TestRoomIssue := structs.RoomIssue{
		BasicRoomInfo: events.BasicRoomInfo{
			BuildingID: "ITB",
			RoomID:     "ITB-1108",
		},
		Alerts:     map[string]structs.Alert{TestAlert.AlertID: TestAlert},
		Severity:   "Critical",
		IncidentID: "INC0432033",
		// Notes:         "After consulting with Xuther, it seems that we should reboot",
		// HelpSentAt:    time.Now(),
		// Responders:    []string{"Joe", "Danny", "John"},
		// HelpArrivedAt: time.Now(),
		Resolved: true,
		ResolutionInfo: structs.ResolutionInfo{
			Code:  "Alert Cleared",
			Notes: "alrerts auto-resolved",
		},
	}

	log.L.Debugf("Test alert %v", TestAlert)

	err := SyncServiceNowWithRoomIssue(TestRoomIssue)

	if err != nil {
		log.L.Debugf("Error: %v", err)
	} else {
		log.L.Debugf("Success!")
	}
}
