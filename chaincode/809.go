package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Point float64 `json:"point"`
	Token float64 `json:"token"`
}

var rate float64

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Name: "Joe", Point: 100, Token: 0},
		{ID: "asset2", Name: "Chalice", Point: 200, Token: 0},
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
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, name string, point float64, token float64) (string, error) {
	exists, _ := s.AssetExists(ctx, id)

	if exists {
		var str = fmt.Sprintf("the asset %s already exists", id)
		return str, nil
	}

	asset := Asset{
		ID:    id,
		Name:  name,
		Point: point,
		Token: token,
	}
	assetJSON, _ := json.Marshal(asset)

	var str = "creat successful"
	return str, ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, name string, point float64, token float64) (string, error) {
	exists, _ := s.AssetExists(ctx, id)

	if !exists {
		var str = fmt.Sprintf("the asset %s does not exist", id)
		return str, nil
	}

	// overwriting original asset with new asset
	asset := Asset{
		ID:    id,
		Name:  name,
		Point: point,
		Token: token,
	}
	assetJSON, _ := json.Marshal(asset)

	var str = "update successful"
	return str, ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes a given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) (string, error) {
	exists, _ := s.AssetExists(ctx, id)

	if !exists {
		var str = fmt.Sprintf("the asset %s does not exist", id)
		return str, nil
	}
	var str = "delete successful"
	return str, ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferPoint Transfer point to token with given id in world state.
func (s *SmartContract) TransferPoint(ctx contractapi.TransactionContextInterface, id string, transpoint float64) (string, error) {
	asset, _ := s.ReadAsset(ctx, id)

	if asset.Point < transpoint {
		var str = "point balance is insufficient"
		return str, nil
	}

	asset.Point -= transpoint
	asset.Token += transpoint * rate

	assetJSON, _ := json.Marshal(asset)

	var str = fmt.Sprintf("transfer successful, now %s %s has %f tokens", asset.Name, asset.ID, asset.Token)
	return str, ctx.GetStub().PutState(id, assetJSON)
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

// GetRate return the rate of transfer now
func (s *SmartContract) GetRate() string {
	str := fmt.Sprintf("The current rate is %f.", rate)
	return str
}

// SetRate change the rate of transfer
func (s *SmartContract) SetRate(newrate float64) string {
	if newrate < 0.0000 {
		var str = "please set the rate to a number greater than 0"
		return str
	}
	rate = newrate
	str := fmt.Sprintf("rate change succeeded, now rate is %f.", rate)
	return str
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}
