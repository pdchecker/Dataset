package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//LoanBuyerObj ...
type LoanBuyerObj struct {
	LoanBuyerID   int       `json:"loanbuyerID"`
	LoanBuyerName string    `json:"loanbuyername"`
	BuyerCategory string    `json:"buyercategory"`
	AdminUserID   int       `json:"adminuserID"`
	HqAddress     string    `json:"hqaddress"`
	Location      string    `json:"location"`
	LocationLat   string    `json:"locationlat"`
	LocationLong  string    `json:"locationlong"`
	Created   time.Time `json:"created"`
	Createdby string    `json:"createdby"`
}

//LoanBuyerHistory ...
type LoanBuyerHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	LoanBuyer *LoanBuyerObj  `json:"loanbuyer"`
}

//LoanBuyerContract for handling writing and reading from the world state
type LoanBuyerContract struct {
	contractapi.Contract
}




//Put adds a new key with value to the world state
func (sc *LoanBuyerContract) Put(ctx contractapi.TransactionContextInterface, loanbuyerID int, loanbuyername string, buyercategory string, adminuserID int, hqaddress string, location string, locationlat string, locationlong string)	(err error) {

	if loanbuyerID == 0 {
		err = errors.New("Loan Buyer ID can not be empty")
		return
	}
	
	obj := new(LoanBuyerObj)
	obj.LoanBuyerID = loanbuyerID
	obj.LoanBuyerName = loanbuyername
	obj.BuyerCategory = buyercategory
	obj.AdminUserID = adminuserID
	obj.HqAddress = hqaddress
	obj.Location = location
	obj.LocationLat = locationlat
	obj.LocationLong = locationlong

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(loanbuyerID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *LoanBuyerContract) Get(ctx contractapi.TransactionContextInterface, key string) (*LoanBuyerObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	loanBuyerObj := new(LoanBuyerObj)
	if err := json.Unmarshal(existingObj, loanBuyerObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type LoanBuyerObj", key)
	}
    return loanBuyerObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *LoanBuyerContract) History(ctx contractapi.TransactionContextInterface, key string) ([]LoanBuyerHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []LoanBuyerHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(LoanBuyerObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := LoanBuyerHistory{
			TxID:		state.GetTxId(),
			Timestamp:	time.Unix(state.GetTimestamp().GetSeconds(), 0),
			LoanBuyer:	entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
