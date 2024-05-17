package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// struct for SmartContract defines the available Car Ownership Smart Contract methods
type CarOwnershipSmartContract struct {
	// embed the contractapi.Contract struct to get access to the Contract's methods
	contractapi.Contract
}

// Car describes basic details
type Car struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Number string `json:"number"`
	Owner  string `json:"owner"`
	Cost   int    `json:"cost"`
}

// Add new car to the ledger
func (cosc *CarOwnershipSmartContract) AddCar(ctx contractapi.TransactionContextInterface, id, name, number, owner string, cost int) error {
	// check if id already exists
	bs, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read id from ledger world state : %s", err.Error())
	}

	if bs != nil {
		return fmt.Errorf("the property %s already exists in ledger", id)
	}

	c := Car{
		ID:     id,
		Name:   name,
		Number: number,
		Owner:  owner,
		Cost:   cost,
	}

	cbs, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, cbs)
}

// returns all the existing cars in the ledger
func (cosc *CarOwnershipSmartContract) ListAllCars(ctx contractapi.TransactionContextInterface) ([]*Car, error) {
	cIter, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer cIter.Close()

	var cs []*Car
	for cIter.HasNext() {
		propertyResponse, err := cIter.Next()
		if err != nil {
			return nil, err
		}

		var c *Car
		err = json.Unmarshal(propertyResponse.Value, &c)
		if err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}

	return cs, nil
}

// get the car by id from the ledger
func (cosc *CarOwnershipSmartContract) GetCarById(ctx contractapi.TransactionContextInterface, id string) (*Car, error) {
	bs, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read id from world state: %s", err.Error())
	}

	if bs == nil {
		return nil, fmt.Errorf("the property %s does not exist", id)
	}

	var c *Car
	err = json.Unmarshal(bs, &c)

	if err != nil {
		return nil, err
	}
	return c, nil
}

// transfers the ownership of the car to the new owner (creates a entry in the ledger with the new owner name)
func (cosc *CarOwnershipSmartContract) ChangeOwner(ctx contractapi.TransactionContextInterface, id string, newOwner string, cost int) error {
	c, err := cosc.GetCarById(ctx, id)
	if err != nil {
		return err
	}

	c.Owner = newOwner
	c.Cost = cost
	bs, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, bs)
}

func main() {
	cc, err := contractapi.NewChaincode(new(CarOwnershipSmartContract))
	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
