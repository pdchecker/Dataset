package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//RoleObj ...
type RoleObj struct {
	RoleID       int       `json:"roleID"`
	UserCategory string    `json:"usercategory"`
	UserType     string    `json:"usertype"`
	RoleName     string    `json:"rolename"`
	RoleDesc     string    `json:"roledesc"`
	Created      time.Time `json:"created"`
	Createdby    string    `json:"createdby"`
}

//RoleHistory ...
type RoleHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	Role *RoleObj  `json:"role"`
}

//RoleContract for handling writing and reading from the world state
type RoleContract struct {
	contractapi.Contract
}

//Put adds a new key with value to the world state
func (sc *RoleContract) Put(ctx contractapi.TransactionContextInterface, roleID int, usercategory string, usertype string, rolename string, roledesc string)	(err error) {

	if roleID == 0 {
		err = errors.New("Role ID can not be empty")
		return
	}
	
	obj := new(RoleObj)
	obj.RoleID = roleID
	obj.UserCategory = usercategory
	obj.UserType = usertype
	obj.RoleName = rolename
	obj.RoleDesc = roledesc

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(roleID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *RoleContract) Get(ctx contractapi.TransactionContextInterface, key string) (*RoleObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	roleObj := new(RoleObj)
	if err := json.Unmarshal(existingObj, roleObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type RoleObj", key)
	}
    return roleObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *RoleContract) History(ctx contractapi.TransactionContextInterface, key string) ([]RoleHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []RoleHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(RoleObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := RoleHistory{
			TxID:		state.GetTxId(),
			Timestamp:	time.Unix(state.GetTimestamp().GetSeconds(), 0),
			Role:	entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
