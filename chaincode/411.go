package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//BuyerObj ...
type BuyerObj struct {
	BuyerID   int       `json:"buyerID"`
	UserID    int       `json:"userID"`
	BuyerType string    `json:"buyertype"`
	Details   string    `json:"details"`
	RegDate   string    `json:"regdate"`
	Created   time.Time `json:"created"`
	Createdby string    `json:"createdby"`
}

//BuyerHistory ...
type BuyerHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	Buyer *BuyerObj  `json:"buyer"`
}

//BuyerContract for handling writing and reading from the world state
type BuyerContract struct {
	contractapi.Contract
}

//Put adds a new key with value to the world state
func (sc *BuyerContract) Put(ctx contractapi.TransactionContextInterface, buyerID int, userID int, buyertype string, details string, regdate string)	(err error) {

	if buyerID == 0 {
		err = errors.New("Buyer ID can not be empty")
		return
	}

	if userID == 0 {
		err = errors.New("User ID can not be empty")
		return
	}

	
	obj := new(BuyerObj)
	obj.BuyerID = buyerID
	obj.UserID = userID
	obj.BuyerType = buyertype
	obj.Details = details
	obj.RegDate = regdate	

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(buyerID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *BuyerContract) Get(ctx contractapi.TransactionContextInterface, key string) (*BuyerObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	buyerObj := new(BuyerObj)
	if err := json.Unmarshal(existingObj, buyerObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type BuyerObj", key)
	}
    return buyerObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *BuyerContract) History(ctx contractapi.TransactionContextInterface, key string) ([]BuyerHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []BuyerHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(BuyerObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := BuyerHistory{
			TxID:		state.GetTxId(),
			Timestamp:	time.Unix(state.GetTimestamp().GetSeconds(), 0),
			Buyer:	entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
