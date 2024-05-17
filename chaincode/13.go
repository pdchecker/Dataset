/*
SPDX-License-Identifier: Apache-2.0
*/

package main
//package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes
// ID -> reference/serial number of the specific car part (ex: "120.47021-XXXXXXXXXXX")
// Car -> models of car using this specific car part (ex: "Fiat 500, Fiat Panda, Fiat Punto")
// Description -> detailt description of the car part (ex: "Symmetric vane; split-core castings; Black E-Coat anti-corrosive coating protects; Double disc ground friction surface")
// Brand -> Brand of the car part (ex: "Centric")
// ProductionDate ->  (ex: "DD/MM/YYYY")
// ProductionLocation -> (ex: "Saint Jose, US")
type Asset struct {
	ID 					string `json:"ID"`
	Car					string `json:"Car"`
	Brand          		string `json:"Brand"`
	ProductionDate      string `json:"ProductionDate"`
	ProductionLocation	string `json:"ProductionLocation"`
	Description         string `json:"Description"`
}

// adding a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "120.47021-15486957423", Car: "Audi Q1, Audi Q2, Audi Q3, Volkswagen Tiguan", Brand: "Volkswagen", ProductionDate: "04/11/2004", ProductionLocation: "Stuttgart, Germany", Description: "Chassis"},
		{ID: "115.15442-68495214587", Car: "Ford F150, Ford F250", Brand: "Ford", ProductionDate: "23/06/2011", ProductionLocation: "Detroit, US", Description: "Drive Train"},
		{ID: "254.51488-54875265847", Car: "Tesla model S, Tesla model 3, Tesla model Y", Brand: "Tesla", ProductionDate: "30/04/2019", ProductionLocation: "Austin, Texas", Description: "Battery"},
		{ID: "151.51847-84956877413", Car: "Toyota Corolla, Toyota rav4, Yahama  MT-15", Brand: "Toyota", ProductionDate: "12/09/2015", ProductionLocation: "Shanghai, China", Description: "Chip"},
		{ID: "58.41684-65184543156", Car: "Volkswagen Golf, Mini cooper S", Brand: "Thyssenkrupp Steering", ProductionDate: "19/02/2021", ProductionLocation: "Liechtenstein, Liechtenstein", Description: "Steering"},
		{ID: "456.56488-56464864115", Car: "Renault 5", Brand: "Renault", ProductionDate: "10/06/1996", ProductionLocation: "Montbéliard, France", Description: "Headlights"},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to init assets. %v", err)
		}
	}

	return nil
}

// CreateAsset -> create and adds new asset to the network.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, car string, brand string, productiondate string, productionlocation string, description string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             	id,
		Car:          		car,
		Brand:           	brand,
		ProductionDate:     productiondate,
		ProductionLocation: productionlocation,
		Description:		description,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset -> returns specific asset stored in the network
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

// UpdateAsset -> updates existing asset in the network.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, car string, brand string, productiondate string, productionlocation string, description string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	asset := Asset{
		ID:             	id,
		Car:          		car,
		Brand:           	brand,
		ProductionDate:     productiondate,
		ProductionLocation: productionlocation,
		Description:		description,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset -> deletes specific asset from the network.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists -> returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GetAllAssets -> returns all assets in the network
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

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}

