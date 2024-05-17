package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type OwnerAsset struct {
	OwnerID   string `json:"OwnerID"`
	OwnerName string `json:"OwnerName"`
	IsActive  bool   `json:"IsActive"`
}

const Prefix = "Owner:"

func GetOwnerID(ctx contractapi.TransactionContextInterface, Name string) (string, error) {
	//returns ownerID to the app that calls it
	owners_list, err := s.GetAllOwners(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}

	for _, owner := range owners_list {
		if owner.OwnerName == Name {
			return owner.OwnerID
		}
	}

	return "", fmt.Errorf("Owner does not exist")

}

func (s *SmartContract) IsOwnerActive(ctx contractapi.TransactionContextInterface, Name string) (string, error) {
	// returns boolean for owner status
	owners_list, err := s.GetAllOwners(ctx)
	if err != nil {
		return "false", fmt.Errorf("failed to read from world state: %v", err)
	}

	var owner *OwnerAsset
	for _, owner := range owners_list {
		if owner.OwnerName == Name {
			return owner.IsActive
		}
	}
	return "false", nil

}

func (s *SmartContract) MakeOwnerActive(ctx contractapi.TransactionContextInterface, Name string) error {
	owners_list, err := s.GetAllOwners(ctx)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	var owner *OwnerAsset
	for _, iteratorVar := range owners_list {
		if iteratorVar.OwnerName == Name {
			owner = iteratorVar
			break
		}
	}
	owner.IsActive = true
	updatedOwnerJSON, err := json.Marshal(owner)
	if err != nil {
		return fmt.Errorf("failed to marshal updated owner: %v", err)
	}

	return ctx.GetStub().PutState(OwnerPrefix+owner.OwnerID, updatedOwnerJSON)
}

func (s *SmartContract) MakeOwnerInactive(ctx contractapi.TransactionContextInterface, Name string) error {
	owners_list, err := s.GetAllOwners(ctx)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	var owner *OwnerAsset
	for _, iteratorVar := range owners_list {
		if iteratorVar.OwnerName == Name {
			owner = iteratorVar
			break
		}
	}
	owner.IsActive = false
	updatedOwnerJSON, err := json.Marshal(owner)
	if err != nil {
		return fmt.Errorf("failed to marshal updated owner: %v", err)
	}

	return ctx.GetStub().PutState(OwnerPrefix+owner.OwnerID, updatedOwnerJSON)
}

func (s *SmartContract) OwnerExistence(ctx contractapi.TransactionContextInterface, OwnerName string) (string, error) {

	// loop through and match based on NAME- in order to generate id using uuid

	owners_list, err := s.GetAllOwners(ctx)
	if err != nil {
		return "false", fmt.Errorf("Failed to get current existing owners %v", err)
	}

	for _, iteratorVar := range owners_list {
		if iteratorVar.OwnerName == OwnerName {
			return "true", nil
		}
	}

	return "false", nil
}

func (s *SmartContract) RegisterOwner(ctx contractapi.TransactionContextInterface, Name string) error {

	// NAME SHOULD BE PASSED AS A PARAMETER
	// id generation happens in this function- on creation of an owner
	ownerexists, err := s.OwnerExistence(ctx, Name) // TODO: FIND OUT HOW TO CHECK FOR OWNER NAME INSTEAD OF ID (gen ID when registering owner- not here)
	if err != nil {
		return err
	}

	// if owner exists already

	if ownerexists == "true" {
		// now there are 2 possible scenarios- active, the other is inactive
		owneractive, err := s.IsOwnerActive(ctx, Name)
		if err != nil {
			return err
		}
		if owneractive == "true" { // if owner is active i.e. existing and active, a statement is returned that the user is already registered
			fmt.Printf("ERRROR : Owner is already registered!")
			return fmt.Errorf("Owner is already registered")
		} else {
			// if the owner is not active, they are made active
			err := s.MakeOwnerActive(ctx, Name)
			if err != nil {
				return fmt.Errorf("error in changing owner's status")
			}
			fmt.Printf("Owner is active")
			return nil
		}
	}

	// OWNER DOES NOT EXIST- So, now we will create a new owner, and register them too

	id := uuid.New().String() // generating it outside because it won't be accessible if generated inside the initialization of owner
	owner := OwnerAsset{
		OwnerName: Name,
		OwnerID:   id,
		IsActive:  true,
	}
	ownerJSON, err := json.Marshal(owner)
	if err != nil {
		return err
	}
	state_err := ctx.GetStub().PutState(OwnerPrefix+id, ownerJSON) // new state added
	fmt.Printf("Owner creation returned : %s\n", state_err)

	return state_err

}

func (s *SmartContract) UnregisterOwner(ctx contractapi.TransactionContextInterface, Name string) error {

	ownerexists, err := s.OwnerExistence(ctx, Name)

	if err != nil {
		return err
	}
	if ownerexists == "true" {
		owneractive, err := s.IsOwnerActive(ctx, Name)
		if err != nil {
			return err
		}
		if owneractive == "true" { // if owner is active i.e. existing and active, the owner is made inactive
			err := s.MakeOwnerInactive(ctx, Name)
			if err != nil {
				return fmt.Errorf("error in changing owner's status")
			}
			fmt.Printf("Owner is active")
			return nil
		} else {
			fmt.Printf("ERRROR : Owner is already unregistered!")
			return fmt.Errorf("Owner is already unregistered")
		}
	}

	return nil
}

func (s *SmartContract) GetAllOwners(ctx contractapi.TransactionContextInterface) ([]*OwnerAsset, error) {
	iteratorVar, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {

		return nil, err

	}

	defer iteratorVar.Close()

	var owners []*OwnerAsset
	var ownerCount = 0

	for iteratorVar.HasNext() {
		queryResponse, err := iteratorVar.Next()
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(queryResponse.Key, OwnerPrefix) {
			var owner OwnerAsset
			err = json.Unmarshal(queryResponse.Value, &owner)
			if err != nil {
				return nil, err
			}
			owners = append(owners, &owner)
			ownerCount++
		}
	}

	if ownerCount > 0 {
		return owners, nil
	} else {
		return []*OwnerAsset{}, nil
	}

}

func (s *SmartContract) DeleteOwner(ctx contractapi.TransactionContextInterface, Name string) error {

	owners_list, err := s.GetAllOwners(ctx)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}

	var owner *OwnerAsset
	for _, iteratorVar := range owners_list {
		if iteratorVar.OwnerName == Name {
			owner = iteratorVar
			break
		}
	}

	delop := ctx.GetStub().DelState(OwnerPrefix + owner.OwnerID)
	fmt.Printf("Message received on deletion: %s", delop)
	return nil

}
