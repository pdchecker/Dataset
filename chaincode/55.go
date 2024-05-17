package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimpleContract contract for handling writing and reading from the world state
type SimpleContract struct {
	contractapi.Contract
}

// Create adds a new key with value to the world state
func (sc *SimpleContract) Create(ctx CustomTransactionContextInterface, key string, value string) error {
	existing := ctx.GetData()

	if existing != nil {
		return fmt.Errorf("Cannot create world state pair with key %s. Already exists", key)
	}

	err := ctx.GetStub().PutState(key, []byte(value))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Update changes the value with key in the world state
func (sc *SimpleContract) Update(ctx CustomTransactionContextInterface, key string, value string) error {
	existing := ctx.GetData()

	if existing == nil {
		return fmt.Errorf("Cannot update world state pair with key %s. Does not exist", key)
	}

	err := ctx.GetStub().PutState(key, []byte(value))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// Read returns the value at key in the world state
func (sc *SimpleContract) Read(ctx CustomTransactionContextInterface, key string) (string, error) {
	existing := ctx.GetData()
	if existing == nil {
		return "", fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
	}

	return string(existing), nil
}

// GetEvaluateTransactions returns functions of SimpleContract not to be tagged as submit
func (sc *SimpleContract) GetEvaluateTransactions() []string {
	return []string{"Read"}
}
