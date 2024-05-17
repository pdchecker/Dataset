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

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ID             string `json:"asset_id"`
	Color          string `json:"color"`
	Size           int    `json:"size"`
	Owner          string `json:"owner"`
	AppraisedValue int    `json:"appraisedValue"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300},
		{ID: "asset2", Color: "red", Size: 5, Owner: "Brad", AppraisedValue: 400},
		{ID: "asset3", Color: "green", Size: 10, Owner: "Jin Soo", AppraisedValue: 500},
		{ID: "asset4", Color: "yellow", Size: 10, Owner: "Max", AppraisedValue: 600},
		{ID: "asset5", Color: "black", Size: 15, Owner: "Adriana", AppraisedValue: 700},
		{ID: "asset6", Color: "white", Size: 15, Owner: "Michel", AppraisedValue: 800},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, rawAssetEnrollRequest string) (string, error) {

	var assetEnrollRequest Asset
	err := json.Unmarshal([]byte(rawAssetEnrollRequest), &assetEnrollRequest)
	if err != nil {
		return "", err
	}

	exists, err := s.AssetExists(ctx, assetEnrollRequest.ID)
	if err != nil {
		return "", err
	}
	if exists {
		return "", fmt.Errorf("the asset %s already exists", assetEnrollRequest.ID)
	}

	result, err := json.Marshal(assetEnrollRequest)
	if err != nil {
		return "", fmt.Errorf("failed json.Marshal: %v", err)
	}

	err = ctx.GetStub().PutState(assetEnrollRequest.ID, result)
	if err != nil {
		return "", err
	}

	return "CreateAsset OK", nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, rawAssetReadRequest string) (*Asset, error) {
	var assetReadRequest Asset
	err := json.Unmarshal([]byte(rawAssetReadRequest), &assetReadRequest)
	if err != nil {
		return nil, err
	}

	assetJSON, err := ctx.GetStub().GetState(assetReadRequest.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", assetReadRequest.ID)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, rawAssetUpdateRequest string) (string, error) {

	var assetUpdateRequest Asset
	err := json.Unmarshal([]byte(rawAssetUpdateRequest), &assetUpdateRequest)
	if err != nil {
		return "", err
	}

	exists, err := s.AssetExists(ctx, assetUpdateRequest.ID)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("the asset %s does not exist", assetUpdateRequest.ID)
	}

	result, err := json.Marshal(assetUpdateRequest)
	if err != nil {
		return "", fmt.Errorf("failed json.Marshal: %v", err)
	}

	err = ctx.GetStub().PutState(assetUpdateRequest.ID, result)
	if err != nil {
		return "", err
	}

	return "UpdateAsset OK", nil
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, rawAssetDeleteRequest string) (string, error) {
	var assetDeleteRequest Asset
	err := json.Unmarshal([]byte(rawAssetDeleteRequest), &assetDeleteRequest)
	if err != nil {
		return "", err
	}

	exists, err := s.AssetExists(ctx, assetDeleteRequest.ID)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("the asset %s does not exist", assetDeleteRequest.ID)
	}

	err = ctx.GetStub().DelState(assetDeleteRequest.ID)
	if err != nil {
		return "", err
	}

	return "CreateAsset OK", nil
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, rawAssetTransferRequest string) (string, error) {
	var assetTransferRequest Asset
	err := json.Unmarshal([]byte(rawAssetTransferRequest), &assetTransferRequest)
	if err != nil {
		return "", err
	}
	exists, err := s.AssetExists(ctx, assetTransferRequest.ID)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("the asset %s does not exist", assetTransferRequest.ID)
	}

	asset, err := s.ReadAsset(ctx, rawAssetTransferRequest)
	if err != nil {
		return "", err
	}

	asset.Owner = assetTransferRequest.Owner
	result, err := json.Marshal(asset)
	if err != nil {
		return "", fmt.Errorf("failed json.Marshal: %v", err)
	}

	err = ctx.GetStub().PutState(assetTransferRequest.ID, result)
	if err != nil {
		return "", err
	}

	return "TransferAsset OK", nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
