package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//BankObj ...
type BankObj struct {
	BankID          int       `json:"bankID"`
	BankName        string    `json:"bankname"`
	HqAddress       string    `json:"hqaddress"`
	BankCategory    string    `json:"bankcategory"`
	BankAdminUserID int       `json:"bankadminuserID"`
	Location        string    `json:"location"`
	LocationLat     string    `json:"locationlat"`
	LocationLong    string    `json:"locationlong"`
	Created         time.Time `json:"created"`
	Createdby       string    `json:"createdby"`
}

//BankHistory ...
type BankHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	Bank *BankObj  `json:"bank"`
}

//BankContract for handling writing and reading from the world state
type BankContract struct {
	contractapi.Contract
}

//Put adds a new key with value to the world state
func (sc *BankContract) Put(ctx contractapi.TransactionContextInterface, bankID int, bankname string, hqaddress string, bankcategory string, bankadminuserID int, location string, locationlat string, locationlong string)	(err error) {

	if bankID == 0 {
		err = errors.New("Bank ID can not be empty")
		return
	}

	if bankname == "" {
		err = errors.New("Bank Name can not be empty")
		return 
	}
	
	obj := new(BankObj)
	obj.BankID = bankID
	obj.BankName = bankname
	obj.HqAddress = hqaddress
	obj.BankCategory = bankcategory
	obj.BankAdminUserID = bankadminuserID
	obj.Location = location
	obj.LocationLat = locationlat
	obj.LocationLong = locationlong

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(bankID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *BankContract) Get(ctx contractapi.TransactionContextInterface, key string) (*BankObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	bankObj := new(BankObj)
	if err := json.Unmarshal(existingObj, bankObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type BankObj", key)
	}
    return bankObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *BankContract) History(ctx contractapi.TransactionContextInterface, key string) ([]BankHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []BankHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(BankObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := BankHistory{
			TxID:      state.GetTxId(),
			Timestamp: time.Unix(state.GetTimestamp().GetSeconds(), 0),
			Bank:     entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
