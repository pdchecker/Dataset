package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//PropertyObj ...
type PropertyObj struct {
	ProID        int       `json:"proID"`
	SellerID     int       `json:"sellerID"`
	ProType      string    `json:"protype"`
	ProName      string    `json:"proname"`
	Desc         string    `json:"desc"`
	Address      string    `json:"address"`
	Location     string    `json:"location"`
	LocationLat  string    `json:"locationlat"`
	LocationLong string    `json:"locationlong"`
	Views        string    `json:"views"`
	ViewerStats  string    `json:"viewerstats"`
	EntryDate    string    `json:"entrydate"`
	ExpiryDate   string    `json:"expirydate"`
	Status       string    `json:"status"`
	Created       time.Time `json:"created"`
	Createdby     string    `json:"createdby"`
}

//PropertyHistory ...
type PropertyHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	Property *PropertyObj  `json:"property"`
}

//PropertyContract for handling writing and reading from the world state
type PropertyContract struct {
	contractapi.Contract
}

//Put adds a new key with value to the world state
func (sc *PropertyContract) Put(ctx contractapi.TransactionContextInterface, proID int, sellerID int, protype string, proname string, desc string, address string, location string, locationlat string, locationlong string, views string, viewerstats string, entrydate string, expirydate string, status string)	(err error) {

	if proID == 0 {
		err = errors.New("Property ID can not be empty")
		return
	}

	obj := new(PropertyObj)
	obj.ProID = proID
	obj.SellerID = sellerID
	obj.ProType = protype
	obj.ProName = proname
	obj.Desc = desc
	obj.Address = address
	obj.Location = location
	obj.LocationLat = locationlat
	obj.LocationLong = locationlong
	obj.Views = views
	obj.ViewerStats = viewerstats
	obj.EntryDate = entrydate
	obj.ExpiryDate = expirydate
	obj.Status = status

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(proID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *PropertyContract) Get(ctx contractapi.TransactionContextInterface, key string) (*PropertyObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	propertyObj := new(PropertyObj)
	if err := json.Unmarshal(existingObj, propertyObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type PropertyObj", key)
	}
    return propertyObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *PropertyContract) History(ctx contractapi.TransactionContextInterface, key string) ([]PropertyHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []PropertyHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(PropertyObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := PropertyHistory{
			TxID:		state.GetTxId(),
			Timestamp:	time.Unix(state.GetTimestamp().GetSeconds(), 0),
			Property:	entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
