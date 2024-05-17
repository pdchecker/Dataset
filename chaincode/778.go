package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ComplexContract contract for handling BasicAssets
type ComplexContract struct {
	contractapi.Contract
}

// NewAsset adds a new basic asset to the world state using id as key
func (s *ComplexContract) NewAsset(ctx CustomTransactionContextInterface, id string, owner Owner, value int) error {
	existing := ctx.GetData()

	if existing != nil {
		return fmt.Errorf("Cannot create new basic asset in world state as key %s already exists", id)
	}

	ba := new(BasicAsset)
	ba.ID = id
	ba.Owner = owner
	ba.Value = value
	ba.SetConditionNew()

	baBytes, _ := json.Marshal(ba)

	err := ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// UpdateOwner changes the ownership of a basic asset and mark it as used
func (cc *ComplexContract) UpdateOwner(ctx CustomTransactionContextInterface, id string, newOwner Owner) error {
	existing := ctx.GetData()

	if existing == nil {
		return fmt.Errorf("Cannot update asset in world state as key %s does not exists", id)
	}

	ba := new(BasicAsset)

	err := json.Unmarshal(existing, ba)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type BasicAsset", id)
	}

	ba.Owner = newOwner
	ba.SetConditionUsed()

	baBytes, _ := json.Marshal(ba)

	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// UpdateValue changes the value of a basic asset to add the value passed
func (cc *ComplexContract) UpdateValue(ctx CustomTransactionContextInterface, id string, valueAdd int) error {
	existing := ctx.GetData()

	if existing == nil {
		return fmt.Errorf("Cannot update asset in world state as key %s does not exist", id)
	}

	ba := new(BasicAsset)

	err := json.Unmarshal(existing, ba)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type BasicAsset", id)
	}

	ba.Value += valueAdd

	baBytes, _ := json.Marshal(ba)

	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// GetAsset returns the basic asset with id given from the world state
func (cc *ComplexContract) GetAsset(ctx CustomTransactionContextInterface, id string) (*BasicAsset, error) {
	existing := ctx.GetData()

	if existing == nil {
		return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", id)
	}

	ba := new(BasicAsset)

	err := json.Unmarshal(existing, ba)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type BasicAsset", id)
	}

	return ba, nil
}

// GetEvaluateTransactions returns functions of ComplexContract not to be tagged as submit
func (cc *ComplexContract) GetEvaluateTransactions() []string {
	return []string{"GetAsset"}
}
