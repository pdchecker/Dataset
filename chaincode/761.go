package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Contract, error) {
	contractJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	if contractJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Contract
	if err := json.Unmarshal(contractJSON, &asset); err != nil {
		return nil, err
	}

	return &asset, nil
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]Contract, error) {
	var cSlice []Contract

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		log.Println("GetStateByRange err:", err)
		return cSlice, err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return cSlice, err
		}

		var c Contract
		err = json.Unmarshal(queryResponse.Value, &c)
		if err != nil {
			log.Println("GetStateByRange next err:", err)
			return cSlice, err
		}
		cSlice = append(cSlice, c)
	}

	return cSlice, nil
}
