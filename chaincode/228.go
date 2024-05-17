package main

import (
	"encoding/json"
	"fmt"

	"github.com/siddhantprateek/tractionx/contracts/models"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type PropertyContract struct {
	contractapi.Contract
}

// PropertyExists returns true when property with given ID exists in world state
func (c *PropertyContract) PropertyExists(
	ctx contractapi.TransactionContextInterface,
	propertyId string) (bool, error) {

	// GetStub: should provide a way to access the stub set by Init/Invoke
	// GetState: returns the value of the specified `key` from the ledger.
	data, err := ctx.GetStub().GetState(propertyId)
	if err != nil {
		return false, err
	}
	return data != nil, nil
}

func (c *PropertyContract) CreateProperty(
	ctx contractapi.TransactionContextInterface,
	propertyId string,
	plotNumber int,
	builder string,
	owner string,
	currentPrice float32,
	sellingPrice float32) error {

	propertyExist, err := c.PropertyExists(ctx, propertyId)
	if err != nil {
		return err
	}
	if propertyExist {
		return fmt.Errorf("the property %s already exist", propertyId)
	}

	property := models.Property{
		ID:           propertyId,
		PlotNumber:   plotNumber,
		Builder:      builder,
		OwnedBy:      owner,
		CurrentPrice: currentPrice,
		SellingPrice: sellingPrice,
	}
	propertyJSON, err := json.Marshal(property)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(propertyId, propertyJSON)
}

func (c *PropertyContract) ReadProperty(
	ctx contractapi.TransactionContextInterface,
	id string,
) (*models.Property, error) {

	propertyJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}

	if propertyJSON == nil {
		return nil, fmt.Errorf("the property %s does not exists", id)
	}

	var property models.Property
	err = json.Unmarshal([]byte(propertyJSON), &property)
	if err != nil {
		return nil, err
	}

	return &property, nil

}

func (c *PropertyContract) TransferProperty(
	ctx contractapi.TransactionContextInterface,
	id string,
	newOwner string) (string, error) {

	property, err := c.ReadProperty(ctx, id)
	if err != nil {
		return "", err
	}

	oldOwner := property.OwnedBy
	property.OwnedBy = newOwner

	propertyJSON, err := json.Marshal(property)
	if err != nil {
		return "", err
	}
	err = ctx.GetStub().PutState(id, propertyJSON)
	if err != nil {
		return "", err
	}
	return oldOwner, nil
}

// GetAllProperty returns all properties found in world state
func (c *PropertyContract) GetAllProperty(
	ctx contractapi.TransactionContextInterface,
) ([]*models.Property, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var properties []*models.Property
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var property models.Property
		err = json.Unmarshal(queryResponse.Value, &property)
		if err != nil {
			return nil, err
		}
		properties = append(properties, &property)
	}

	return properties, nil
}

// DeleteProperty deletes an given property from the world state.
func (c *PropertyContract) DeleteProperty(
	ctx contractapi.TransactionContextInterface,
	propertyId string) error {
	propertyExist, err := c.PropertyExists(ctx, propertyId)
	if err != nil {
		return err
	}
	if !propertyExist {
		return fmt.Errorf("the property %s does not exist", propertyId)
	}
	return ctx.GetStub().DelState(propertyId)
}

func (c *PropertyContract) UpdateProperty(
	ctx contractapi.TransactionContextInterface,
	propertyId string,
	plotNumber int,
	builder string,
	owner string,
	currentPrice float32,
	sellingPrice float32) error {

	propertyExist, err := c.PropertyExists(ctx, propertyId)
	if err != nil {
		return err
	}
	if !propertyExist {
		return fmt.Errorf("the property %s does not exist", propertyId)
	}

	property := models.Property{
		ID:           propertyId,
		PlotNumber:   plotNumber,
		Builder:      builder,
		OwnedBy:      owner,
		CurrentPrice: currentPrice,
		SellingPrice: sellingPrice,
	}

	propertyJSON, err := json.Marshal(property)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(propertyId, propertyJSON)
}
