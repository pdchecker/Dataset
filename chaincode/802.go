package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//SellerObj ...
type SellerObj struct {
	SellerID   int       `json:"sellerID"`
	UserID     int       `json:"userID"`
	SellerType string    `json:"sellertype"`
	Details    string    `json:"details"`
	RegDate    string    `json:"regdate"`
	Created    time.Time `json:"created"`
	Createdby  string    `json:"createdby"`
}

//SellerHistory ...
type SellerHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	Seller *SellerObj  `json:"seller"`
}

//SellerContract for handling writing and reading from the world state
type SellerContract struct {
	contractapi.Contract
}

//Put adds a new key with value to the world state
func (sc *SellerContract) Put(ctx contractapi.TransactionContextInterface, sellerID int, userID int, sellertype string, details string, regdate string)	(err error) {

	if sellerID == 0 {
		err = errors.New("Seller ID can not be empty")
		return
	}
	
	obj := new(SellerObj)
	obj.SellerID = sellerID
	obj.UserID = userID
	obj.SellerType = sellertype
	obj.Details = details
	obj.RegDate = regdate

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(sellerID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *SellerContract) Get(ctx contractapi.TransactionContextInterface, key string) (*SellerObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	sellerObj := new(SellerObj)
	if err := json.Unmarshal(existingObj, sellerObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type SellerObj", key)
	}
    return sellerObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *SellerContract) History(ctx contractapi.TransactionContextInterface, key string) ([]SellerHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []SellerHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(SellerObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := SellerHistory{
			TxID:		state.GetTxId(),
			Timestamp:	time.Unix(state.GetTimestamp().GetSeconds(), 0),
			Seller:	entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
