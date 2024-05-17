package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Contract chaincode that defines the business logic for managing land

type Contract struct {
	contractapi.Contract
}

// Instantiae does nothing
func (c *Contract) Instantiae() {
	fmt.Println("Instantiated")

}

// Register registers new land and stores it in the world state
func (c *Contract) Register(ctx TransactionContextInterface, owner string, landNumber string, registerDateTime string, faceValue int) (*Land, error) {
	land := Land{LandNumber: landNumber, Owner: owner, RegisterDateTime: registerDateTime, FaceValue: faceValue}

	land.SetIssued()

	err := ctx.GetLandList().AddLand(&land)

	if err != nil {
		return nil, err
	}
	return &land, nil
}

// Sell gets the land ready for market. It sets the state of land in trading, which means other people can make offers to buy land.
//TODO: BETTER NAMING HERE.
func (c *Contract) PutOnSell(ctx TransactionContextInterface, land Land) (*Land, error) {
	land.SetTrading()
	err := ctx.GetLandList().AddLand(&land)

	if err != nil {
		return nil, err
	}
	return &land, nil
}

// ReadAsset returns the asset stored in the world state given the owner and landNumber.
func (contract *Contract) ReadLand(ctx TransactionContextInterface, owner string, landNumber string) (*Land, error) {

	land, err := ctx.GetLandList().GetLand(owner, landNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	return land, nil
}

// Transfer Asset(Land) to a new owner
func (s *Contract) TransferLand(ctx TransactionContextInterface, currentOwner string, landNumber string, newOwner string) error {
	land, err := s.ReadLand(ctx, currentOwner, landNumber)
	if err != nil {
		return err
	}

	if land.Owner != currentOwner {
		return fmt.Errorf("Land %s is not owned by %s", landNumber, currentOwner)
	}
	if !land.IsTrading() {
		return fmt.Errorf("Land %s is not for sale", landNumber)
	}

	land.Owner = newOwner
	return ctx.GetLandList().UpdateLand(land)

}

// GetAllAssets returns all assets found in world state
func (contract *Contract) GetAllLand(ctx contractapi.TransactionContextInterface) ([]*Land, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Land
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var land Land
		err = json.Unmarshal(queryResponse.Value, &land)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &land)
	}

	return assets, nil
}

// AssetExists returns true when asset with given (owner and landNumber) exists in world state
func (contract *Contract) LandExists(ctx contractapi.TransactionContextInterface, owner string, landNubmer string) (bool, error) {
	landJSON, err := ctx.GetStub().GetState(CreateLandKey(owner, landNubmer))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return landJSON != nil, nil
}
