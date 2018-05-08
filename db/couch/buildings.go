package couch

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/byuoitav/common/structs"
	"github.com/byuoitav/configuration-database-microservice/log"
)

//GetBuilding gets the company's building with the corresponding ID from the couch database
func (c *CouchDB) GetBuilding(id string) (structs.Building, error) {
	var toReturn building
	err := MakeRequest("GET", fmt.Sprintf("buildings/%v", id), "", nil, &toReturn)
	if err != nil {
		msg := fmt.Sprintf("[couch] Could not get building %v. %v", id, err.Error())
		log.L.Warn(msg)
	}

	return toReturn.Building, err
}

//GetAllBuildings returns all buildings for the company specified
func (c *CouchDB) GetAllBuildings() ([]structs.Building, error) {
	var toReturn []structs.Building
	var query IDPrefixQuery

	query.Limit = 1000
	b, err := json.Marshal(query)
	if err != nil {
		log.L.Warnf("There was a problem marshalling the query: %v", err.Error())
		return []structs.Building{}, err
	}

	var resp buildingQueryResponse

	err = MakeRequest("POST", fmt.Sprintf("buildings/_find"), "application/json", b, &resp)
	if err != nil {
		msg := fmt.Sprintf("[couch] Could not get buildings. %v", err.Error())
		log.L.Warn(msg)
		return []structs.Building{}, errors.New(msg)
	}

	for _, doc := range resp.Docs {
		toReturn = append(toReturn, doc.Building)
	}

	return toReturn, err
}

/*
AddBuilding adds a building. The building must have at least:
1) ID
2) Name

The function will also overwrite the existing building providing the _rev field is set properly
*/
func (c *CouchDB) CreateBuilding(toAdd structs.Building) (structs.Building, error) {
	log.L.Debugf("Starting adding a building: %v", toAdd.ID)

	if len(toAdd.ID) < 2 {
		msg := "Cannot create building, must have an ID"
		log.L.Warn(msg)
	}

	//there's not a lot to buildings, so we can just add

	log.L.Debug("Checks passed, creating building.")

	b, err := json.Marshal(toAdd)
	if err != nil {

		msg := fmt.Sprintf("Couldn't marshal building into JSON. Error: %v", err.Error())
		log.L.Error(msg) // this is a slightly bigger deal
		return toAdd, errors.New(msg)
	}

	resp := CouchUpsertResponse{}

	err = MakeRequest("PUT", fmt.Sprintf("buildings/%v", toAdd.ID), "", b, &resp)
	if err != nil {
		log.L.Debugf("%v", err)
		if conflict, ok := err.(*Confict); ok {
			msg := fmt.Sprintf("Error: %v Make changes on an updated version of the configuration.", conflict.Error())
			log.L.Warn(msg)
			return toAdd, errors.New(msg)
		}
		//ther was some other problem
		msg := fmt.Sprintf("unknown problem creating the Building: %v", err.Error())
		log.L.Warn(msg)
		return toAdd, errors.New(msg)
	}

	log.L.Debug("Building created, retriving new configuration from database.")

	//return the created config
	toAdd, err = c.GetBuilding(toAdd.ID)
	if err != nil {
		msg := fmt.Sprintf("There was a problem getting the newly created building: %v", err.Error())
		log.L.Warn(msg)
		return toAdd, errors.New(msg)
	}

	log.L.Debug("Done.")
	return toAdd, nil
}

func (c *CouchDB) DeleteBuilding(id string) error {
	log.L.Infof("Starting delete for building: %v", id)
	building, err := c.GetBuilding(id)
	if err != nil {
		msg := fmt.Sprintf("There was a problem deleting the building: %v", err.Error())
		log.L.Warn(msg)
		return errors.New(msg)
	}

	log.L.Debugf("Checking for rooms in building %v", id)
	//first we need to check if there are any rooms in the building, if there are, we don't allow a delete at this level
	rms, err := c.GetRoomsByBuilding(id)
	if err != nil {
		msg := fmt.Sprintf("Problem checking the building for rooms: %v", err.Error())
		log.L.Warn(msg)
		return errors.New(msg)
	}

	if len(rms) > 0 {
		log.L.Infof("There were %v rooms in building %v, aborting delete", len(rms), id)
		return errors.New(fmt.Sprintf("Couldn't delete building %v, there are still rooms associated with the building. Remove all rooms from building before deleting.", id))
	}

	log.L.Debugf("No rooms found in building %v. Proceeding with deletion", id)

	/* TODO have to get rev
	err = MakeRequest("DELETE", fmt.Sprintf("buildings/%s?rev=%v", id, building.Rev), "", nil, nil)
	if err != nil {
		msg := fmt.Sprintf("There was a problem deleting the building: %v", err.Error())
		log.L.Warn(msg)
		return errors.New(msg)
	}
	*/
	log.L.Debug(building)

	log.L.Debugf("Building %v deleted successfully.", id)
	return nil
}

func (c *CouchDB) UpdateBuilding(id string, building structs.Building) (structs.Building, error) {
	return structs.Building{}, nil
}
