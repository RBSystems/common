package couch

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/byuoitav/common/structs"
)

func (c *CouchDB) GetBuilding(id string) (structs.Building, error) {
	resp, err := c.getBuilding(id)
	return *resp.Building, err
}

func (c *CouchDB) getBuilding(id string) (building, error) {
	var toReturn building

	err := c.MakeRequest("GET", fmt.Sprintf("%v/%v", BUILDINGS, id), "", nil, &toReturn)
	if err != nil {
		err = &NotFound{msg: fmt.Sprintf("failed to get building %v: %s", id, err)}
	}

	return toReturn, err
}

func (c *CouchDB) GetAllBuildings() ([]structs.Building, error) {
	var toReturn []structs.Building
	var query IDPrefixQuery
	query.Selector.ID.GT = "\x00"
	query.Limit = 1000

	b, err := json.Marshal(query)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to marshal query to get all buildings: %s", err))
	}

	var resp buildingQueryResponse

	err = c.MakeRequest("POST", fmt.Sprintf("%v/_find", BUILDINGS), "application/json", b, &resp)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to get all buildings: %s", err))
	}

	for _, doc := range resp.Docs {
		toReturn = append(toReturn, *doc.Building)
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
	var toReturn structs.Building

	// validate building
	err := toAdd.Validate()
	if err != nil {
		return toReturn, err
	}

	b, err := json.Marshal(toAdd)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("failed to marshal building %s: %s", toAdd.ID, err))
	}

	// post new building
	var resp CouchUpsertResponse
	err = c.MakeRequest("POST", BUILDINGS, "", b, &resp)
	if err != nil {
		// if there was a conflict
		if conflict, ok := err.(*Conflict); ok {
			return toReturn, errors.New(fmt.Sprintf("building already exists, please update this building or change id's. error: %s", conflict))
		}

		// or an unknown error
		return toReturn, errors.New(fmt.Sprintf("unable to create building %s: %s", toAdd.ID, err))
	}

	// return the created building
	toAdd, err = c.GetBuilding(toAdd.ID)
	if err != nil {
		return toReturn, errors.New(fmt.Sprintf("unable getting the building %s after creating it: %s", toAdd.ID, err))
	}

	return toReturn, nil
}

func (c *CouchDB) DeleteBuilding(id string) error {
	// get the rev of the building to delete
	building, err := c.getBuilding(id)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to get building %s to delete: %s", id, err))
	}

	// check if there are any rooms in the building; if there are, don't allow deletion
	rms, err := c.GetRoomsByBuilding(id)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to check the building for rooms: %s", err))
	}

	if len(rms) > 0 {
		return errors.New(fmt.Sprintf("there are still rooms associated with the building %s. delete all rooms from it first.", id))
	}

	// make request to delete building
	err = c.MakeRequest("DELETE", fmt.Sprintf("%s/%s?rev=%v", BUILDINGS, id, building.Rev), "", nil, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to delete building %s: %s", id, err))
	}

	return nil
}

// delete a building without checking if rooms will be affected
func (c *CouchDB) deleteBuildingWithoutCascade(id string) error {
	building, err := c.getBuilding(id)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to get building %s to delete: %s", id, err))
	}

	err = c.MakeRequest("DELETE", fmt.Sprintf("%s/%s?rev=%v", BUILDINGS, id, building.Rev), "", nil, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to delete building %s: %s", id, err))
	}

	return nil
}

func (c *CouchDB) UpdateBuilding(id string, building structs.Building) (structs.Building, error) {
	var toReturn structs.Building

	// validate updated building
	err := building.Validate()
	if err != nil {
		return toReturn, err
	}

	if id == building.ID { // the building ID isn't changing
		// get the rev of the building
		bld, err := c.getBuilding(id)
		if err != nil {
			return toReturn, errors.New(fmt.Sprintf("unable to get building %s to update: %s", err))
		}

		// marshal the new building
		b, err := json.Marshal(building)
		if err != nil {
			return toReturn, errors.New(fmt.Sprintf("unable to unmarshal new building: %s", err))
		}

		// update the building
		err = c.MakeRequest("PUT", fmt.Sprintf("%s/%s?rev=%v", BUILDINGS, id, bld.Rev), "application/json", b, &toReturn)
		if err != nil {
			return toReturn, errors.New(fmt.Sprintf("failed to update building %s: %s", id, err))
		}
	} else { // the builiding ID is changing :|
		// delete the old building
		err = c.deleteBuildingWithoutCascade(id)
		if err != nil {
			return toReturn, errors.New(fmt.Sprintf("unable to delete old building %s: %s", id, err))
		}

		// create the new building
		_, err = c.CreateBuilding(building)
		if err != nil {
			return toReturn, errors.New(fmt.Sprintf("unable to create new building %s: %s", id, err))
		}

		// get rooms that were in the old building
		rooms, err := c.GetRoomsByBuilding(id)
		if err != nil {
			return toReturn, errors.New(fmt.Sprintf("unable to get rooms assocated with old building: %s"))
		}

		// update each of the rooms to be in the new building
		for _, room := range rooms {
			go func() {
				// create the new room id
				oldID := room.ID
				room.ID = strings.Replace(room.ID, id, building.ID, 1)

				// update the room
				c.UpdateRoom(oldID, room)
			}()
		}
	}

	return toReturn, nil
}
