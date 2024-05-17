/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// AssetContract contract for managing CRUD for Asset
type AssetContract struct {
	contractapi.Contract
}

// AssetExists returns true when asset with given ID exists in world state
func (c *AssetContract) AssetExists(ctx contractapi.TransactionContextInterface, assetID string) (bool, error) {
	data, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// CreateAsset creates a new instance of Asset
func (c *AssetContract) CreateAsset(ctx contractapi.TransactionContextInterface, assetID string, name string, price int) error {
	exists, err := c.AssetExists(ctx, assetID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if exists {
		return fmt.Errorf("The asset %s already exists", assetID)
	}

	asset := new(Asset)
	asset.Name = name
	asset.Price = price

	bytes, _ := json.Marshal(asset)

	return ctx.GetStub().PutState(assetID, bytes)
}

// ReadAsset retrieves an instance of Asset from the world state
func (c *AssetContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {
	exists, err := c.AssetExists(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("The asset %s does not exist", assetID)
	}

	bytes, _ := ctx.GetStub().GetState(assetID)

	asset := new(Asset)

	err = json.Unmarshal(bytes, asset)

	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal world state data to type Asset")
	}

	return asset, nil
}

// UpdateAsset retrieves an instance of Asset from the world state and updates its name and price
func (c *AssetContract) UpdateAsset(ctx contractapi.TransactionContextInterface, assetID string, newName string, newPrice int) error {
	exists, err := c.AssetExists(ctx, assetID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", assetID)
	}

	asset := new(Asset)
	asset.Name = newName
	asset.Price = newPrice

	bytes, _ := json.Marshal(asset)

	return ctx.GetStub().PutState(assetID, bytes)
}

// DeleteAsset deletes an instance of Asset from the world state
func (c *AssetContract) DeleteAsset(ctx contractapi.TransactionContextInterface, assetID string) error {
	exists, err := c.AssetExists(ctx, assetID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", assetID)
	}

	return ctx.GetStub().DelState(assetID)
}
