package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type User struct {
	Name   string `json:"name"`
	PubKey string `json:"pubkey"`
}

func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, name string) (bool, error) {
	userJSON, err := ctx.GetStub().GetState(name)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return userJSON != nil, nil
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, name string, pubkey string) error {
	exists, err := s.UserExists(ctx, name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the user %s already exists", name)
	}

	user := User{
		Name:   name,
		PubKey: pubkey,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(name, userJSON)
}

// ReadUser returns the asset stored in the world state with given id.
func (s *SmartContract) ReadUser(ctx contractapi.TransactionContextInterface, name string) (*User, error) {
	userJSON, err := ctx.GetStub().GetState(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", name)
	}

	var user User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAllUsers returns all assets found in world state
func (s *SmartContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*User, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var users []*User
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}
