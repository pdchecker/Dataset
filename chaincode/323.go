package main

import (
	"encoding/json"
	"fmt"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Contract
type SmartContract struct {
	contractapi.Contract
}

// InitLedger is just a template for now.
// Used to test the connection and verify that applications can connect to the chaincode.
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	return nil
}

func (s *SmartContract) CreateContract(ctx contractapi.TransactionContextInterface, contractJSON string) error {
	var part lib.Part
	err := json.Unmarshal([]byte(contractJSON), &part)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	exists, err := s.ContractExists(ctx, part.Timestamp)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the Contract %v already exists", part.Timestamp)
	}

	return ctx.GetStub().PutState(part.Timestamp, []byte(contractJSON))
}

// ContractExists returns true when Contract with given ID exists in world state
func (s *SmartContract) ContractExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	ContractJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %w", err)
	}

	return ContractJSON != nil, nil
}

func (s *SmartContract) GetAssetByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]lib.Part, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []lib.Part
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset lib.Part
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

func (s *SmartContract) GetAssetQualityByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]lib.Quality, error) {
	assets, err := s.GetAssetByRange(ctx, startKey, endKey)
	if err != nil {
		return nil, err
	}

	var qualities = make([]lib.Quality, 1)

	for _, asset := range assets {
		if asset.DocumentBody.Quality == 1 {
			qualities[0].High += 1
		} else {
			qualities[0].Low += 1
		}
	}
	qualities[0].Total = len(assets)
	return qualities, nil
}
