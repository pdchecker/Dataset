/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a car
type SmartContract struct {
	contractapi.Contract
}

// Car describes basic details of what makes up a car
type Firmware struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Path    string `json:"path"`
	Create  string `json:"create"`
	Deploy  string `json:"deploy"`
	Hash    string `json:"hash"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Firmware
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// CreateCar adds a new car to the world state with given details
func (s *SmartContract) CreateFirmware(ctx contractapi.TransactionContextInterface, firmwareNUmber string, id string, name string, version string, status string, path string, create string, deploy string, hash string) error {
	firmware := Firmware{
		ID:      id,
		Name:    name,
		Version: version,
		Status:  status,
		Path:    path,
		Create:  create,
		Deploy:  deploy,
		Hash:    hash,
	}

	firmwareAsBytes, _ := json.Marshal(firmware)

	return ctx.GetStub().PutState(firmwareNUmber, firmwareAsBytes)
}

// QueryCar returns the car stored in the world state with given id
func (s *SmartContract) QueryFirmware(ctx contractapi.TransactionContextInterface, firmwareNumber string) (*Firmware, error) {
	firmwareAsBytes, err := ctx.GetStub().GetState(firmwareNumber)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if firmwareAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", firmwareNumber)
	}

	firmware := new(Firmware)
	_ = json.Unmarshal(firmwareAsBytes, firmware)

	return firmware, nil
}

// QueryAllCars returns all cars found in world state
func (s *SmartContract) QueryAllFirmwares(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		firmware := new(Firmware)
		_ = json.Unmarshal(queryResponse.Value, firmware)

		queryResult := QueryResult{Key: queryResponse.Key, Record: firmware}
		results = append(results, queryResult)
	}

	return results, nil
}

// ChangeCarOwner updates the owner field of car with given id in world state
func (s *SmartContract) ChangeFirmwareVersion(ctx contractapi.TransactionContextInterface, firmwareNumber string, newVersion string) error {
	firmware, err := s.QueryFirmware(ctx, firmwareNumber)

	if err != nil {
		return err
	}

	firmware.Version = newVersion

	firmwareAsBytes, _ := json.Marshal(firmware)

	return ctx.GetStub().PutState(firmwareNumber, firmwareAsBytes)
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
