package servicenow

import (
	"fmt"
	"net/http"

	"github.com/byuoitav/common"
	"github.com/byuoitav/common/jsonhttp"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/structs"
)

var token = "***REMOVED***"

func main() {
	log.SetLevel("debug")
	port := ":8025"
	router := common.NewRouter()
	//Create incident test
	// TestAlert := structs.Alert{
	// 	BuildingID: "ITB",
	// 	RoomID:     "1108",
	// 	DeviceID:   "ITB-1108-CP5",
	// 	Message:    "There is an issue with the pi, it is not turning on",
	// 	Data:       "Stuff",
	// }
	// Createincident(TestAlert)

	//Modify incident test
	// ModifyAlert := structs.Alert{
	// 	HelpSentAt:    time.Now(),
	// 	HelpArrivedAt: time.Now().Add(5),
	// }
	// SysID := "89233ae61bdb674003e68622dd4bcb1b"
	// ModifyIncident(SysID, ModifyAlert)

	//query incident resolution category (for closing tickets) test
	// table := "u_inc_resolution_cat"
	// PrintIncidentResolutionCategory(table)

	//close incident test
	// sysID := "89233ae61bdb674003e68622dd4bcb1b"
	// resolutionaction := "Replaced"
	// notes := "I replaced the pi and the room is working now"
	// CloseIncident(sysID, resolutionaction, notes)

	//test Query all incidents for AV-Support
	// GroupName := "AV-Support"
	// QueryIncidentsByGroup(GroupName)

	//test query by room
	BuildingID := "ITB"
	RoomID := "1108"
	QueryIncidentsByRoom(BuildingID, RoomID)
	server := http.Server{
		Addr:           port,
		MaxHeaderBytes: 1024 * 10,
	}
	router.StartServer(&server)

}

func Createincident(Alert structs.Alert) (structs.Incident, error) {

	weburl := "https://api.byu.edu/domains/servicenow/incident/v1.1/incident"
	room := fmt.Sprintf("%s %s", Alert.BuildingID, Alert.RoomID)
	workStatus := "Very Low"
	sensitivity := "Very Low"
	severity := "Very Low"
	reach := "Very Low"
	assignmentGroup := "AV-Support"
	shortDescription := fmt.Sprintf("%s in room %s-%s has the following alert: %s.", Alert.DeviceID, Alert.BuildingID, Alert.RoomID, Alert.Message)
	description := fmt.Sprintf("%s in room %s-%s has the following alert: %s.", Alert.DeviceID, Alert.BuildingID, Alert.RoomID, Alert.Message)
	internalNotes := fmt.Sprintf("%v", Alert.Data)
	input := structs.Incident{
		Service:          "TEC Room",
		Room:             room,
		WorkStatus:       workStatus,
		Sensitivity:      sensitivity,
		Severity:         severity,
		Reach:            reach,
		AssignmentGroup:  assignmentGroup,
		ShortDescription: shortDescription,
		Description:      description,
		InternalNotes:    internalNotes,
		CallerId:         "mjsmith3",
	}
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	var output structs.IncidentWrapper
	outputJson, _, err := jsonhttp.CreateAndExecuteJSONRequest("CreateRequest", "POST", weburl,
		input, headers, 20, &output)
	log.L.Debugf("Output JSON: %s", outputJson)
	log.L.Debugf("Output JSON: %+v", output)
	return output.Result, err
}

//we need to be able to access the sysID of the incident Ticket
//TO DO: takes incident ID and string for internal notes
func ModifyIncident(SysID string, Alert structs.Alert) (structs.ReceiveIncident, error) {
	weburl := fmt.Sprintf("https://api.byu.edu/domains/servicenow/incident/v1.1/incident/%s?sysparm_display_value=true", SysID)
	log.L.Debugf("WebURL: %s", weburl)
	var internalNotes string
	var state string

	if Alert.HelpSentAt.IsZero() {

	} else {
		internalNotes = "Help was was sent at: " + fmt.Sprintf("%s", Alert.HelpSentAt)
		state = "Assigned"
	}

	if Alert.HelpArrivedAt.IsZero() {

	} else {
		internalNotes += "\n" + " Help arrived at: " + fmt.Sprintf("%s", Alert.HelpArrivedAt)
		state = "Work In Progress"
	}

	input := structs.Incident{
		State:         state,
		InternalNotes: internalNotes,
		Description:   "This is a description, want to see what happens",
	}
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	var output structs.ReceiveIncidentWrapper
	outputJson, _, err := jsonhttp.CreateAndExecuteJSONRequest("ModifyIncident", "PUT", weburl,
		input, headers, 20, &output)
	log.L.Debugf("Output JSON: %s", outputJson)
	log.L.Debugf("Output JSON: %+v", output)
	return output.Result, err

}

func GetResolutionActions() (structs.ResolutionCategories, error) {
	weburl := "https://api.byu.edu:443/domains/servicenow/tableapi/v1/table/u_inc_resolution_cat?sysparm_query=active%3Dtrue%5Eassignment_group%3Djavascript%3AgetMyAssignmentGroups()"
	log.L.Debugf("WebURL: %s", weburl)
	var output structs.ResolutionCategories
	input := ""
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}
	outputJson, _, err := jsonhttp.CreateAndExecuteJSONRequest("querycategory", "GET", weburl,
		input, headers, 20, &output)
	log.L.Debugf("Output JSON: %s", outputJson)
	log.L.Debugf("Output JSON: %+v", output)
	return output, err
}

func CloseIncident(SysID string, resolutionaction string, internalNotes string) (structs.ReceiveIncident, error) {
	weburl := fmt.Sprintf("https://api.byu.edu/domains/servicenow/incident/v1.1/incident/%s?sysparm_display_value=true", SysID)
	log.L.Debugf("WebURL: %s", weburl)
	state := "Closed"
	closurecode := "Resolved"
	resolutionservice := "TEC Room"

	input := structs.Incident{
		State:             state,
		InternalNotes:     internalNotes,
		ClosureCode:       closurecode,
		ResolutionService: resolutionservice,
		ResolutionAction:  resolutionaction,
	}
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}

	var output structs.ReceiveIncidentWrapper
	outputJson, _, err := jsonhttp.CreateAndExecuteJSONRequest("ModifyIncident", "PUT", weburl,
		input, headers, 20, &output)
	log.L.Debugf("Output JSON: %s", outputJson)
	log.L.Debugf("Output JSON: %+v", output)
	return output.Result, err
}

//query all incidents for a given assignment group
func QueryIncidentsByGroup(GroupName string) (structs.QueriedIncidents, error) {
	weburl := fmt.Sprintf("https://api.byu.edu/domains/servicenow/incident/v1.1/incident?active=true&assignment_group=%s&sysparm_display_value=true", GroupName)
	log.L.Debugf("WebURL: %s", weburl)
	var output structs.QueriedIncidents
	input := ""
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}
	outputJson, _, err := jsonhttp.CreateAndExecuteJSONRequest("querycategory", "GET", weburl,
		input, headers, 20, &output)
	log.L.Debugf("Output JSON: %s", outputJson)
	log.L.Debugf("Output JSON: %+v", output)
	return output, err
}

//query all incidents by room number
func QueryIncidentsByRoom(BuildingID string, RoomID string) (structs.QueriedIncidents, error) {
	weburl := fmt.Sprintf("https://api.byu.edu/domains/servicenow/incident/v1.1/incident?active=true&sysparm_display_value=true&u_room=%s+%s", BuildingID, RoomID)
	log.L.Debugf("WebURL: %s", weburl)
	var output structs.QueriedIncidents
	input := ""
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}
	outputJson, _, err := jsonhttp.CreateAndExecuteJSONRequest("querycategory", "GET", weburl,
		input, headers, 20, &output)
	log.L.Debugf("Output JSON: %s", outputJson)
	log.L.Debugf("Output JSON: %+v", output)
	return output, err
}

//TODO query all of the users in the system (Net_id)
//Get ticket
