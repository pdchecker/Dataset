package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//BranchConfigObj ...
type BranchConfigObj struct {
	ConfigID		int       `json:"configID"`
	BankID			int       `json:"bankID"`
	ConfigName		string    `json:"configname"`
	ConfigDesc		string    `json:"configdesc"`
	Item			string    `json:"item"`
	Value			string    `json:"value"`
	Created			time.Time `json:"created"`
	Createdby		string    `json:"createdby"`
}

//BranchConfigHistory ...
type BranchConfigHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	BranchConfig *BranchConfigObj  `json:"branchconfig"`
}

//BranchConfigContract for handling writing and reading from the world state
type BranchConfigContract struct {
	contractapi.Contract
}

//Put adds a new key with value to the world state
func (sc *BranchConfigContract) Put(ctx contractapi.TransactionContextInterface, configID int, bankID int, configname string, configdesc string, item string, value string)	(err error) {

	if configID == 0 {
		err = errors.New("Config ID can not be empty")
		return
	}

	if bankID == 0 {
		err = errors.New("Bank ID can not be empty")
		return
	}

	
	obj := new(BranchConfigObj)
	obj.ConfigID = configID
	obj.BankID = bankID
	obj.ConfigName = configname
	obj.ConfigDesc = configdesc

	obj.Value = value
	obj.Item = item
	obj.ConfigDesc = configdesc

	

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(configID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *BranchConfigContract) Get(ctx contractapi.TransactionContextInterface, key string) (*BranchConfigObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	branchConfigObj := new(BranchConfigObj)
	if err := json.Unmarshal(existingObj, branchConfigObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type BranchConfigObj", key)
	}
    return branchConfigObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *BranchConfigContract) History(ctx contractapi.TransactionContextInterface, key string) ([]BranchConfigHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []BranchConfigHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(BranchConfigObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := BranchConfigHistory{
			TxID:		state.GetTxId(),
			Timestamp:	time.Unix(state.GetTimestamp().GetSeconds(), 0),
			BranchConfig:	entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
